package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"
)

type RecommendationNumber struct {
	Red        []int   `json:"red"`
	Blue       []int   `json:"blue"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type RecommendationResult struct {
	Summary string                 `json:"summary"`
	Basis   string                 `json:"basis"`
	Numbers []RecommendationNumber `json:"numbers"`
}

type RecommendationProvider interface {
	Generate(ctx context.Context, definition Definition, lotteryType model.LotteryType, history []model.DrawResult, count int) (*RecommendationResult, error)
}

type mockRecommendationProvider struct{}

type openAIRecommendationProvider struct{}

func newRecommendationProvider(provider string) RecommendationProvider {
	if provider == ProviderOpenAICompatible {
		return &openAIRecommendationProvider{}
	}
	return &mockRecommendationProvider{}
}

func (provider *mockRecommendationProvider) Generate(_ context.Context, definition Definition, _ model.LotteryType, history []model.DrawResult, count int) (*RecommendationResult, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	redScores := make(map[int]int)
	blueScores := make(map[int]int)

	for _, draw := range history {
		for _, red := range parseCSVNumbers(draw.RedNumbers) {
			redScores[red]++
		}
		for _, blue := range parseCSVNumbers(draw.BlueNumbers) {
			blueScores[blue]++
		}
	}

	numbers := make([]RecommendationNumber, 0, count)
	for index := 0; index < count; index++ {
		numbers = append(numbers, RecommendationNumber{
			Red:        weightedPick(redScores, definition.RedMin, definition.RedMax, definition.RedCount, rng),
			Blue:       weightedPick(blueScores, definition.BlueMin, definition.BlueMax, definition.BlueCount, rng),
			Confidence: 0.45 + float64(index)*0.05,
			Reason:     "基于近期开奖频次和随机扰动生成，适合作为保底推荐",
		})
	}

	return &RecommendationResult{
		Summary: "使用历史频次加权生成推荐",
		Basis:   fmt.Sprintf("样本期数 %d，红蓝球分别按近期出现频率加权抽样", len(history)),
		Numbers: numbers,
	}, nil
}

func (provider *openAIRecommendationProvider) Generate(ctx context.Context, definition Definition, _ model.LotteryType, history []model.DrawResult, count int) (*RecommendationResult, error) {
	content, err := callRecommendationModel(ctx, definition.Recommendation.Model, buildRecommendationPrompt(definition, history, count))
	if err != nil {
		return nil, err
	}

	result := RecommendationResult{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	if len(result.Numbers) == 0 {
		return nil, fmt.Errorf("AI 未返回推荐号码")
	}
	return &result, nil
}

func buildRecommendationPrompt(definition Definition, history []model.DrawResult, count int) string {
	lines := make([]string, 0, len(history))
	for _, draw := range history {
		lines = append(lines, fmt.Sprintf("%s: 红球[%s] 蓝球[%s]", draw.Issue, draw.RedNumbers, draw.BlueNumbers))
	}

	if definition.Recommendation.Prompt != "" {
		return fmt.Sprintf(
			"%s\n规则：红球 %d 个(%d-%d)，蓝球 %d 个(%d-%d)。\n最近开奖如下：\n%s\n请只返回 JSON：{\"summary\":\"\",\"basis\":\"\",\"numbers\":[{\"red\":[1,2,3,4,5,6],\"blue\":[7],\"confidence\":0.8,\"reason\":\"\"}]}",
			definition.Recommendation.Prompt,
			definition.RedCount,
			definition.RedMin,
			definition.RedMax,
			definition.BlueCount,
			definition.BlueMin,
			definition.BlueMax,
			strings.Join(lines, "\n"),
		)
	}

	return fmt.Sprintf(
		"请为%s生成 %d 组推荐号码。规则：红球 %d 个(%d-%d)，蓝球 %d 个(%d-%d)。最近开奖如下：\n%s\n请只返回 JSON：{\"summary\":\"\",\"basis\":\"\",\"numbers\":[{\"red\":[1,2,3,4,5,6],\"blue\":[7],\"confidence\":0.8,\"reason\":\"\"}]}",
		definition.Name,
		count,
		definition.RedCount,
		definition.RedMin,
		definition.RedMax,
		definition.BlueCount,
		definition.BlueMin,
		definition.BlueMax,
		strings.Join(lines, "\n"),
	)
}

func weightedPick(scores map[int]int, minValue int, maxValue int, count int, rng *rand.Rand) []int {
	selected := make([]int, 0, count)
	used := make(map[int]struct{}, count)

	for len(selected) < count {
		totalWeight := 0
		for number := minValue; number <= maxValue; number++ {
			if _, ok := used[number]; ok {
				continue
			}
			totalWeight += max(1, scores[number])
		}

		target := rng.Intn(max(1, totalWeight))
		accumulator := 0
		chosen := minValue
		for number := minValue; number <= maxValue; number++ {
			if _, ok := used[number]; ok {
				continue
			}
			accumulator += max(1, scores[number])
			if accumulator > target {
				chosen = number
				break
			}
		}
		selected = append(selected, chosen)
		used[chosen] = struct{}{}
	}

	return parseCSVNumbers(formatNumbers(selected))
}

func GenerateRecommendation(ctx context.Context, code string, count int) (*model.Recommendation, error) {
	lotteryType, err := getLotteryType(code)
	if err != nil {
		return nil, err
	}
	definition, err := GetDefinition(code)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		count = max(1, definition.Recommendation.Count)
	}

	historyWindow := max(10, definition.Recommendation.HistoryWindow)
	history := make([]model.DrawResult, 0)
	if err := db.DB.Where("lottery_code = ?", code).Order("issue desc").Limit(historyWindow).Find(&history).Error; err != nil {
		return nil, err
	}

	provider := newRecommendationProvider(resolveValue(definition.Recommendation.Provider, ProviderMock))
	result, err := provider.Generate(ctx, definition, lotteryType, history, count)
	if err != nil && definition.Recommendation.Provider == ProviderOpenAICompatible {
		result, err = newRecommendationProvider(ProviderMock).Generate(ctx, definition, lotteryType, history, count)
	}
	if err != nil {
		return nil, err
	}

	recommendation := model.Recommendation{
		LotteryCode:   code,
		Issue:         guessNextIssue(history),
		Provider:      resolveValue(definition.Recommendation.Provider, ProviderMock),
		Model:         resolveValue(definition.Recommendation.Model, "history-weighted-mock"),
		Strategy:      "history+ai",
		PromptVersion: definition.Recommendation.PromptVersion,
		Summary:       result.Summary,
		Basis:         result.Basis,
		RawPayload:    mustJSON(result),
	}
	if err := db.DB.Create(&recommendation).Error; err != nil {
		return nil, err
	}

	entries := make([]model.RecommendationEntry, 0, len(result.Numbers))
	for index, item := range result.Numbers {
		entries = append(entries, model.RecommendationEntry{
			RecommendationID: recommendation.Id,
			Sequence:         index + 1,
			RedNumbers:       formatNumbers(item.Red),
			BlueNumbers:      formatNumbers(item.Blue),
			Confidence:       item.Confidence,
			Reason:           item.Reason,
		})
	}
	if len(entries) > 0 {
		if err := db.DB.Create(&entries).Error; err != nil {
			return nil, err
		}
	}

	if err := db.DB.Preload("Entries").First(&recommendation, "id = ?", recommendation.Id).Error; err != nil {
		return nil, err
	}
	return &recommendation, nil
}

func guessNextIssue(history []model.DrawResult) string {
	if len(history) == 0 {
		return ""
	}
	current, err := strconv.Atoi(history[0].Issue)
	if err != nil {
		return history[0].Issue
	}
	return strconv.Itoa(current + 1)
}
