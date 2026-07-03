package lottery

import (
	"context"
	"fmt"
	"strings"
	"testing"

	model "go-fiber-starter/internal/model/lottery"
)

func TestBuildRecommendationPromptUsesDefinitionSchema(t *testing.T) {
	definition := Definition{
		Code:      "dlt",
		Name:      "体彩大乐透",
		RedCount:  5,
		RedMin:    1,
		RedMax:    35,
		BlueCount: 2,
		BlueMin:   1,
		BlueMax:   12,
		Recommendation: RecommendationSettings{
			Prompt: "你是大乐透推荐助手。",
		},
	}

	prompt := buildRecommendationPrompt(definition, nil, RecommendationBlocklist{
		RecentRecommendations: []string{"红球[01,02,03,04,05] 蓝球[01,02]"},
	}, 2)

	if !strings.Contains(prompt, "\"red\":[1,2,3,4,5]") {
		t.Fatalf("prompt should use DLT red count example: %s", prompt)
	}
	if !strings.Contains(prompt, "\"blue\":[1,2]") {
		t.Fatalf("prompt should use DLT blue count example: %s", prompt)
	}
	if strings.Contains(prompt, "\"red\":[1,2,3,4,5,6]") {
		t.Fatalf("prompt should not use SSQ red count example: %s", prompt)
	}
	if !strings.Contains(prompt, "近期已推荐") {
		t.Fatalf("prompt should include recent recommendation blocklist: %s", prompt)
	}
}

func TestNormalizeRecommendationNumbersFiltersDuplicatesAndBlocklist(t *testing.T) {
	definition := Definition{
		Code:      "ssq",
		Name:      "福彩双色球",
		RedCount:  6,
		RedMin:    1,
		RedMax:    33,
		BlueCount: 1,
		BlueMin:   1,
		BlueMax:   16,
	}
	blocklist := RecommendationBlocklist{
		RecentRecommendations: []string{
			formatRecommendationSignature("01,02,03,04,05,06", "07"),
		},
		HistoryDraws: []string{
			formatRecommendationSignature("02,04,06,08,10,12", "14"),
		},
	}
	items := []RecommendationNumber{
		{Red: []int{1, 2, 3, 4, 5, 6}, Blue: []int{7}},
		{Red: []int{2, 4, 6, 8, 10, 12}, Blue: []int{14}},
		{Red: []int{3, 3, 8, 12, 21, 30}, Blue: []int{9}},
		{Red: []int{9, 1, 28, 16, 22, 31}, Blue: []int{6}, Confidence: 0.81},
		{Red: []int{1, 9, 16, 22, 28, 31}, Blue: []int{6}, Confidence: 0.82},
		{Red: []int{4, 11, 18, 23, 27, 32}, Blue: []int{15}},
	}

	result, err := normalizeRecommendationNumbers(definition, items, 2, blocklist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("unexpected result count: %d", len(result))
	}
	if got := recommendationNumberSignature(result[0]); got != "红球[01,09,16,22,28,31] 蓝球[06]" {
		t.Fatalf("unexpected first signature: %s", got)
	}
	if got := recommendationNumberSignature(result[1]); got != "红球[04,11,18,23,27,32] 蓝球[15]" {
		t.Fatalf("unexpected second signature: %s", got)
	}
}

func TestBuildRecommendationBlocklistIncludesHistoryDraws(t *testing.T) {
	blocklist := buildRecommendationBlocklistFromHistory([]model.DrawResult{
		{RedNumbers: "01,02,03,04,05,06", BlueNumbers: "07"},
	})
	if len(blocklist.HistoryDraws) != 1 {
		t.Fatalf("unexpected history draw count: %d", len(blocklist.HistoryDraws))
	}
	if blocklist.HistoryDraws[0] != "红球[01,02,03,04,05,06] 蓝球[07]" {
		t.Fatalf("unexpected history signature: %s", blocklist.HistoryDraws[0])
	}
}

func TestRecommendationGenerateCompensatesInvalidJSON(t *testing.T) {
	restoreRecommendationModelCaller(t)
	calls := 0
	recommendationModelCaller = func(ctx context.Context, model string, prompt string) (string, error) {
		calls++
		if calls == 1 {
			return "这不是 JSON", nil
		}
		if !strings.Contains(prompt, "补偿修复要求") {
			t.Fatalf("expected compensation prompt, got: %s", prompt)
		}
		return `{"summary":"ok","basis":"retry","numbers":[{"red":[1,9,16,22,28,31],"blue":[6],"confidence":0.8,"reason":"补偿生成"},{"red":[4,11,18,23,27,32],"blue":[15],"confidence":0.7,"reason":"补偿生成"}]}`, nil
	}

	result, err := (&openAIRecommendationProvider{}).Generate(
		context.Background(),
		testSSQRecommendationDefinition(),
		model.LotteryType{},
		nil,
		RecommendationBlocklist{},
		2,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("unexpected call count: %d", calls)
	}
	if len(result.Numbers) != 2 {
		t.Fatalf("unexpected recommendation count: %d", len(result.Numbers))
	}
}

func TestRecommendationGenerateCompensatesInsufficientCandidates(t *testing.T) {
	restoreRecommendationModelCaller(t)
	calls := 0
	recommendationModelCaller = func(ctx context.Context, model string, prompt string) (string, error) {
		calls++
		if calls == 1 {
			return `{"summary":"bad","basis":"重复","numbers":[{"red":[1,2,3,4,5,6],"blue":[7],"confidence":0.8,"reason":"重复"}]}`, nil
		}
		if !strings.Contains(prompt, "候选不足") {
			t.Fatalf("expected failure reason in compensation prompt, got: %s", prompt)
		}
		return `{"summary":"ok","basis":"retry","numbers":[{"red":[1,9,16,22,28,31],"blue":[6],"confidence":0.8,"reason":"补偿生成"}]}`, nil
	}

	result, err := (&openAIRecommendationProvider{}).Generate(
		context.Background(),
		testSSQRecommendationDefinition(),
		model.LotteryType{},
		nil,
		RecommendationBlocklist{
			RecentRecommendations: []string{
				formatRecommendationSignature("01,02,03,04,05,06", "07"),
			},
		},
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("unexpected call count: %d", calls)
	}
	if recommendationNumberSignature(result.Numbers[0]) != "红球[01,09,16,22,28,31] 蓝球[06]" {
		t.Fatalf("unexpected recommendation: %s", recommendationNumberSignature(result.Numbers[0]))
	}
}

func TestRecommendationGenerateDoesNotRetryConfigError(t *testing.T) {
	restoreRecommendationModelCaller(t)
	calls := 0
	recommendationModelCaller = func(ctx context.Context, model string, prompt string) (string, error) {
		calls++
		return "", fmt.Errorf("未配置 OpenAI 兼容模型")
	}

	_, err := (&openAIRecommendationProvider{}).Generate(
		context.Background(),
		testSSQRecommendationDefinition(),
		model.LotteryType{},
		nil,
		RecommendationBlocklist{},
		1,
	)
	if err == nil {
		t.Fatalf("expected error")
	}
	if calls != 1 {
		t.Fatalf("config error should not retry, got calls: %d", calls)
	}
	if !strings.Contains(err.Error(), "AI 模型连接配置不完整") {
		t.Fatalf("expected config failure reason, got: %v", err)
	}
	if !strings.Contains(err.Error(), "已尝试 1 次") {
		t.Fatalf("expected actual attempt count, got: %v", err)
	}
}

func restoreRecommendationModelCaller(t *testing.T) {
	t.Helper()
	original := recommendationModelCaller
	t.Cleanup(func() {
		recommendationModelCaller = original
	})
}

func testSSQRecommendationDefinition() Definition {
	return Definition{
		Code:      "ssq",
		Name:      "福彩双色球",
		RedCount:  6,
		RedMin:    1,
		RedMax:    33,
		BlueCount: 1,
		BlueMin:   1,
		BlueMax:   16,
		Recommendation: RecommendationSettings{
			Model:  "test-model",
			Prompt: "你是双色球推荐助手。",
		},
	}
}
