package lottery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
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

type openAIRecommendationProvider struct{}

func newRecommendationProvider(provider string) RecommendationProvider {
	if provider != ProviderOpenAICompatible {
		return nil
	}
	return &openAIRecommendationProvider{}
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
	result.Numbers = normalizeRecommendationNumbers(result.Numbers, count)
	if len(result.Numbers) == 0 {
		return nil, fmt.Errorf("AI 返回的推荐号码为空")
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
			"%s\n请严格生成 %d 组推荐号码，不能多也不能少。\n规则：红球 %d 个(%d-%d)，蓝球 %d 个(%d-%d)。\n最近开奖如下：\n%s\n请只返回 JSON：{\"summary\":\"\",\"basis\":\"\",\"numbers\":[{\"red\":[1,2,3,4,5,6],\"blue\":[7],\"confidence\":0.8,\"reason\":\"\"}]}，其中 numbers 数组长度必须等于 %d。",
			definition.Recommendation.Prompt,
			count,
			definition.RedCount,
			definition.RedMin,
			definition.RedMax,
			definition.BlueCount,
			definition.BlueMin,
			definition.BlueMax,
			strings.Join(lines, "\n"),
			count,
		)
	}

	return fmt.Sprintf(
		"请为%s生成 %d 组推荐号码，不能多也不能少。规则：红球 %d 个(%d-%d)，蓝球 %d 个(%d-%d)。最近开奖如下：\n%s\n请只返回 JSON：{\"summary\":\"\",\"basis\":\"\",\"numbers\":[{\"red\":[1,2,3,4,5,6],\"blue\":[7],\"confidence\":0.8,\"reason\":\"\"}]}，其中 numbers 数组长度必须等于 %d。",
		definition.Name,
		count,
		definition.RedCount,
		definition.RedMin,
		definition.RedMax,
		definition.BlueCount,
		definition.BlueMin,
		definition.BlueMax,
		strings.Join(lines, "\n"),
		count,
	)
}

func normalizeRecommendationNumbers(items []RecommendationNumber, count int) []RecommendationNumber {
	if count <= 0 || len(items) <= count {
		return items
	}
	return items[:count]
}

func GenerateRecommendation(ctx context.Context, code string, count int, userID string) (*model.Recommendation, error) {
	userUUID, err := parseRequiredUserID(userID)
	if err != nil {
		return nil, err
	}

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
	if err := db.DB.Where("lottery_code = ?", code).Order("draw_date desc").Order("issue desc").Limit(historyWindow).Find(&history).Error; err != nil {
		return nil, err
	}

	if definition.Recommendation.Model == "" {
		return nil, fmt.Errorf("%s 未配置推荐模型", definition.Name)
	}
	targetIssue, targetDrawDate, err := buildRecommendationPlan(definition, history, time.Now())
	if err != nil {
		return nil, err
	}
	providerName := resolveValue(definition.Recommendation.Provider, ProviderOpenAICompatible)
	provider := newRecommendationProvider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("未配置可用的推荐模型提供方")
	}
	result, err := provider.Generate(ctx, definition, lotteryType, history, count)
	if err != nil {
		return nil, err
	}

	recommendation := model.Recommendation{}
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		saveModeCreate := false
		findErr := tx.Where("user_id = ? AND lottery_code = ? AND issue = ?", userUUID, code, targetIssue).Order("created_at desc").First(&recommendation).Error
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			recommendation = model.Recommendation{
				UserID:      &userUUID,
				LotteryCode: code,
				Issue:       targetIssue,
			}
			saveModeCreate = true
		} else if findErr != nil {
			return findErr
		}

		applyRecommendationSnapshot := func() {
			recommendation.DrawDate = &targetDrawDate
			recommendation.UserID = &userUUID
			recommendation.Provider = providerName
			recommendation.Model = definition.Recommendation.Model
			recommendation.Strategy = "history+ai"
			recommendation.PromptVersion = definition.Recommendation.PromptVersion
			recommendation.Summary = result.Summary
			recommendation.Basis = result.Basis
			recommendation.RawPayload = mustJSON(result)
			recommendation.CheckedAt = nil
			recommendation.PrizeAmount = 0
		}
		applyRecommendationSnapshot()

		if saveModeCreate {
			if createErr := tx.Create(&recommendation).Error; createErr != nil {
				if !isUniqueConstraintError(createErr) {
					return createErr
				}
				if retryErr := tx.Where("user_id = ? AND lottery_code = ? AND issue = ?", userUUID, code, targetIssue).Order("created_at desc").First(&recommendation).Error; retryErr != nil {
					return retryErr
				}
				applyRecommendationSnapshot()
				if saveErr := tx.Omit("Entries").Save(&recommendation).Error; saveErr != nil {
					return saveErr
				}
			}
		} else {
			if err := tx.Omit("Entries").Save(&recommendation).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("recommendation_id = ?", recommendation.Id).Delete(&model.RecommendationEntry{}).Error; err != nil {
			return err
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
		if len(entries) == 0 {
			return nil
		}
		return tx.Create(&entries).Error
	}); err != nil {
		return nil, err
	}

	if err := db.DB.Preload("Entries").First(&recommendation, "id = ?", recommendation.Id).Error; err != nil {
		return nil, err
	}
	return &recommendation, nil
}
