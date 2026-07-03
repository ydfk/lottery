package lottery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"

	"github.com/google/uuid"
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

type RecommendationBlocklist struct {
	RecentRecommendations []string
	HistoryDraws          []string
}

type RecommendationProvider interface {
	Generate(ctx context.Context, definition Definition, lotteryType model.LotteryType, history []model.DrawResult, blocklist RecommendationBlocklist, count int) (*RecommendationResult, error)
}

type openAIRecommendationProvider struct{}

const recommendationGenerateMaxAttempts = 3

var recommendationModelCaller = callRecommendationModel

func newRecommendationProvider(provider string) RecommendationProvider {
	if provider != ProviderOpenAICompatible {
		return nil
	}
	return &openAIRecommendationProvider{}
}

func (provider *openAIRecommendationProvider) Generate(ctx context.Context, definition Definition, _ model.LotteryType, history []model.DrawResult, blocklist RecommendationBlocklist, count int) (*RecommendationResult, error) {
	prompt := buildRecommendationPrompt(definition, history, blocklist, count)
	var lastErr error
	var lastContent string
	attempted := 0
	for attempt := 1; attempt <= recommendationGenerateMaxAttempts; attempt++ {
		attempted = attempt
		result, content, err := provider.generateOnce(ctx, definition, prompt, blocklist, count)
		if err == nil {
			if attempt > 1 {
				logger.Info("AI 推荐生成第 %d 次补偿成功", attempt)
			}
			return result, nil
		}

		lastErr = err
		lastContent = content
		if !isRepairableRecommendationError(err) || attempt == recommendationGenerateMaxAttempts {
			break
		}

		logger.Warn("AI 推荐生成第 %d 次失败，将补偿重试: %s", attempt, detectRecommendationFailureCause(err))
		prompt = buildRecommendationCompensationPrompt(definition, history, blocklist, count, attempt, err, content)
	}

	return nil, fmt.Errorf(
		"AI 推荐生成失败，已尝试 %d 次，最后原因：%s，响应片段：%s: %w",
		attempted,
		detectRecommendationFailureCause(lastErr),
		formatRecommendationResponsePreview(lastContent),
		lastErr,
	)
}

func (provider *openAIRecommendationProvider) generateOnce(ctx context.Context, definition Definition, prompt string, blocklist RecommendationBlocklist, count int) (*RecommendationResult, string, error) {
	content, err := recommendationModelCaller(ctx, definition.Recommendation.Model, prompt)
	if err != nil {
		return nil, "", err
	}

	result, err := parseRecommendationResult(content)
	if err != nil {
		return nil, content, err
	}
	result.Numbers, err = normalizeRecommendationNumbers(definition, result.Numbers, count, blocklist)
	if err != nil {
		return nil, content, err
	}
	return result, content, nil
}

func parseRecommendationResult(content string) (*RecommendationResult, error) {
	result := RecommendationResult{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("AI 返回内容不是合法 JSON: %w", err)
	}
	if len(result.Numbers) == 0 {
		return nil, fmt.Errorf("AI 未返回推荐号码")
	}
	return &result, nil
}

func buildRecommendationPrompt(definition Definition, history []model.DrawResult, blocklist RecommendationBlocklist, count int) string {
	return buildRecommendationPromptWithCandidateCount(
		definition,
		history,
		blocklist,
		count,
		recommendationCandidateCount(count),
	)
}

func buildRecommendationPromptWithCandidateCount(definition Definition, history []model.DrawResult, blocklist RecommendationBlocklist, count int, candidateCount int) string {
	lines := make([]string, 0, len(history))
	for _, draw := range history {
		lines = append(lines, fmt.Sprintf("%s: 红球[%s] 蓝球[%s]", draw.Issue, draw.RedNumbers, draw.BlueNumbers))
	}

	basePrompt := strings.TrimSpace(definition.Recommendation.Prompt)
	if basePrompt == "" {
		basePrompt = fmt.Sprintf("你是%s推荐助手。请结合最近开奖历史，为用户输出结构化 JSON 推荐结果。", definition.Name)
	}

	return fmt.Sprintf(
		"%s\n\n请生成 %d 组候选号码，系统会从中筛选 %d 组最终推荐。\n硬性规则：\n1. 红球必须恰好 %d 个，范围 %d-%d，组内不能重复。\n2. 蓝球必须恰好 %d 个，范围 %d-%d，组内不能重复。\n3. numbers 内不能出现完全相同的红球+蓝球组合。\n4. 不要与“近期已推荐”和“最近开奖”中的完整组合重复。\n5. 不要使用 01,02,03 这类机械连续模板作为主要策略，需兼顾奇偶、大小区间、冷热号和分散度。\n6. reason 用一句话说明选号依据，confidence 使用 0.55-0.95 之间的小数。\n最近开奖：\n%s\n\n近期已推荐：\n%s\n\n请只返回 JSON：%s，其中 numbers 数组长度必须等于 %d。",
		basePrompt,
		candidateCount,
		count,
		definition.RedCount,
		definition.RedMin,
		definition.RedMax,
		definition.BlueCount,
		definition.BlueMin,
		definition.BlueMax,
		formatPromptList(lines),
		formatPromptList(blocklist.RecentRecommendations),
		buildRecommendationJSONExample(definition),
		candidateCount,
	)
}

func buildRecommendationCompensationPrompt(definition Definition, history []model.DrawResult, blocklist RecommendationBlocklist, count int, attempt int, lastErr error, lastContent string) string {
	candidateCount := recommendationCompensationCandidateCount(count, attempt)
	basePrompt := buildRecommendationPromptWithCandidateCount(definition, history, blocklist, count, candidateCount)
	return fmt.Sprintf(
		"%s\n\n补偿修复要求：\n上一次生成失败原因：%s。\n上一次响应片段：%s。\n请针对失败原因修补：如果是 JSON 问题，只输出一个完整 JSON 对象；如果是候选不足、重复、越界或数量错误，请重新生成 %d 组全新候选，并主动避开所有已列出的完整组合。不要解释，不要 Markdown，不要代码块。",
		basePrompt,
		detectRecommendationFailureCause(lastErr),
		formatRecommendationResponsePreview(lastContent),
		candidateCount,
	)
}

func recommendationCandidateCount(count int) int {
	if count <= 0 {
		count = 1
	}
	candidateCount := max(count+4, count*4)
	if candidateCount > 50 {
		return 50
	}
	return candidateCount
}

func recommendationCompensationCandidateCount(count int, attempt int) int {
	if count <= 0 {
		count = 1
	}
	candidateCount := recommendationCandidateCount(count) + count*max(1, attempt)*2
	if candidateCount > 80 {
		return 80
	}
	return candidateCount
}

func formatPromptList(items []string) string {
	if len(items) == 0 {
		return "无"
	}
	return strings.Join(items, "\n")
}

func buildRecommendationJSONExample(definition Definition) string {
	red := make([]int, 0, definition.RedCount)
	for number := definition.RedMin; number <= definition.RedMax && len(red) < definition.RedCount; number++ {
		red = append(red, number)
	}
	blue := make([]int, 0, definition.BlueCount)
	for number := definition.BlueMin; number <= definition.BlueMax && len(blue) < definition.BlueCount; number++ {
		blue = append(blue, number)
	}

	return mustJSON(RecommendationResult{
		Summary: "简要说明本次推荐策略",
		Basis:   "说明参考的历史窗口、冷热号和分布依据",
		Numbers: []RecommendationNumber{
			{
				Red:        red,
				Blue:       blue,
				Confidence: 0.72,
				Reason:     "示例：冷热搭配，区间分布均衡",
			},
		},
	})
}

func normalizeRecommendationNumbers(definition Definition, items []RecommendationNumber, count int, blocklist RecommendationBlocklist) ([]RecommendationNumber, error) {
	if count <= 0 {
		count = 1
	}

	blocked := buildRecommendationSignatureSet(blocklist)
	seen := make(map[string]struct{}, len(items))
	result := make([]RecommendationNumber, 0, count)
	failureCounts := make(map[string]int)
	for _, item := range items {
		normalized, err := normalizeRecommendationNumber(definition, item)
		if err != nil {
			failureCounts[err.Error()]++
			continue
		}

		signature := recommendationNumberSignature(normalized)
		if _, exists := blocked[signature]; exists {
			failureCounts["与近期推荐或历史开奖完整重复"]++
			continue
		}
		if _, exists := seen[signature]; exists {
			failureCounts["候选内完整组合重复"]++
			continue
		}

		seen[signature] = struct{}{}
		result = append(result, normalized)
		if len(result) == count {
			return result, nil
		}
	}

	return nil, fmt.Errorf(
		"AI 推荐候选不足：需要 %d 组，筛选后仅 %d 组，过滤原因：%s",
		count,
		len(result),
		formatRecommendationFailureCounts(failureCounts),
	)
}

func normalizeRecommendationNumber(definition Definition, item RecommendationNumber) (RecommendationNumber, error) {
	if len(item.Red) != definition.RedCount {
		return RecommendationNumber{}, fmt.Errorf("红球数量不正确")
	}
	if len(item.Blue) != definition.BlueCount {
		return RecommendationNumber{}, fmt.Errorf("蓝球数量不正确")
	}
	if containsDuplicate(item.Red) {
		return RecommendationNumber{}, fmt.Errorf("红球号码不能重复")
	}
	if containsDuplicate(item.Blue) {
		return RecommendationNumber{}, fmt.Errorf("蓝球号码不能重复")
	}
	for _, number := range item.Red {
		if number < definition.RedMin || number > definition.RedMax {
			return RecommendationNumber{}, fmt.Errorf("红球号码超出范围")
		}
	}
	for _, number := range item.Blue {
		if number < definition.BlueMin || number > definition.BlueMax {
			return RecommendationNumber{}, fmt.Errorf("蓝球号码超出范围")
		}
	}

	normalized := item
	normalized.Red = append([]int(nil), item.Red...)
	normalized.Blue = append([]int(nil), item.Blue...)
	slices.Sort(normalized.Red)
	slices.Sort(normalized.Blue)
	normalized.Confidence = normalizeRecommendationConfidence(item.Confidence)
	normalized.Reason = strings.TrimSpace(item.Reason)
	return normalized, nil
}

func normalizeRecommendationConfidence(value float64) float64 {
	if value <= 0 {
		return 0.6
	}
	if value > 1 {
		return 1
	}
	return value
}

func buildRecommendationSignatureSet(blocklist RecommendationBlocklist) map[string]struct{} {
	result := make(map[string]struct{}, len(blocklist.RecentRecommendations)+len(blocklist.HistoryDraws))
	for _, signature := range blocklist.RecentRecommendations {
		result[signature] = struct{}{}
	}
	for _, signature := range blocklist.HistoryDraws {
		result[signature] = struct{}{}
	}
	return result
}

func recommendationNumberSignature(item RecommendationNumber) string {
	return formatRecommendationSignature(formatNumbers(item.Red), formatNumbers(item.Blue))
}

func formatRecommendationSignature(redNumbers string, blueNumbers string) string {
	return fmt.Sprintf("红球[%s] 蓝球[%s]", redNumbers, blueNumbers)
}

func formatRecommendationFailureCounts(items map[string]int) string {
	if len(items) == 0 {
		return "没有足够候选"
	}

	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s %d 组", key, items[key]))
	}
	return strings.Join(parts, "；")
}

func detectRecommendationFailureCause(err error) string {
	if err == nil {
		return "未知原因"
	}
	if errors.Is(err, context.Canceled) {
		return "请求已取消"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "模型请求超时"
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "未配置 OpenAI 兼容模型"):
		return "AI 模型连接配置不完整"
	case strings.Contains(message, "HTTP 401") || strings.Contains(message, "HTTP 403"):
		return "AI 接口鉴权失败"
	case strings.Contains(message, "HTTP 400") || strings.Contains(message, "HTTP 404"):
		return "AI 请求参数或模型地址可能不兼容"
	case strings.Contains(message, "HTTP 429"):
		return "AI 接口限流"
	case strings.Contains(message, "HTTP 5"):
		return "AI 服务端临时错误"
	case strings.Contains(message, "不是合法 JSON"):
		return "模型没有返回合法 JSON"
	case strings.Contains(message, "未返回推荐号码"):
		return "模型返回 JSON 中缺少 numbers 候选"
	case strings.Contains(message, "候选不足"):
		return message
	default:
		return message
	}
}

func isRepairableRecommendationError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}

	message := err.Error()
	if strings.Contains(message, "未配置 OpenAI 兼容模型") ||
		strings.Contains(message, "HTTP 400") ||
		strings.Contains(message, "HTTP 401") ||
		strings.Contains(message, "HTTP 404") ||
		strings.Contains(message, "HTTP 403") {
		return false
	}
	return true
}

func formatRecommendationResponsePreview(content string) string {
	preview := strings.TrimSpace(content)
	if preview == "" {
		return "<empty>"
	}
	return truncateLogValue(preview)
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
	targetIssue, targetDrawDate, err := planRecommendationTarget(definition, history, time.Now())
	if err != nil {
		return nil, err
	}
	existing, err := findExistingRecommendation(userUUID, code, targetIssue)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	providerName := resolveValue(definition.Recommendation.Provider, ProviderOpenAICompatible)
	provider := newRecommendationProvider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("未配置可用的推荐模型提供方")
	}
	blocklist, err := buildRecommendationBlocklist(userUUID, code, history)
	if err != nil {
		return nil, err
	}
	result, err := provider.Generate(ctx, definition, lotteryType, history, blocklist, count)
	if err != nil {
		return nil, err
	}

	recommendation := model.Recommendation{}
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if existing, err := findExistingRecommendationWithDB(tx, userUUID, code, targetIssue); err != nil {
			return err
		} else if existing != nil {
			recommendation = *existing
			return nil
		}

		recommendation = model.Recommendation{
			UserID:        &userUUID,
			LotteryCode:   code,
			Issue:         targetIssue,
			DrawDate:      &targetDrawDate,
			Provider:      providerName,
			Model:         definition.Recommendation.Model,
			Strategy:      "history+ai",
			PromptVersion: definition.Recommendation.PromptVersion,
			Summary:       result.Summary,
			Basis:         result.Basis,
			RawPayload:    mustJSON(result),
			CheckedAt:     nil,
			PrizeAmount:   0,
		}
		if err := tx.Create(&recommendation).Error; err != nil {
			if isUniqueConstraintError(err) {
				existing, queryErr := findExistingRecommendationWithDB(tx, userUUID, code, targetIssue)
				if queryErr != nil {
					return queryErr
				}
				if existing != nil {
					recommendation = *existing
					return nil
				}
			}
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

func findExistingRecommendation(userID uuid.UUID, code string, issue string) (*model.Recommendation, error) {
	return findExistingRecommendationWithDB(db.DB, userID, code, issue)
}

func findExistingRecommendationWithDB(database *gorm.DB, userID uuid.UUID, code string, issue string) (*model.Recommendation, error) {
	recommendation := model.Recommendation{}
	err := database.Preload("Entries").
		Where("user_id = ? AND lottery_code = ? AND issue = ?", userID, code, issue).
		Order("created_at asc").
		Order("id asc").
		First(&recommendation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &recommendation, nil
}

func buildRecommendationBlocklist(userID uuid.UUID, code string, history []model.DrawResult) (RecommendationBlocklist, error) {
	recent, err := loadRecentRecommendationSignatures(userID, code, 20)
	if err != nil {
		return RecommendationBlocklist{}, err
	}

	blocklist := buildRecommendationBlocklistFromHistory(history)
	blocklist.RecentRecommendations = recent
	return blocklist, nil
}

func buildRecommendationBlocklistFromHistory(history []model.DrawResult) RecommendationBlocklist {
	historyDraws := make([]string, 0, len(history))
	for _, draw := range history {
		historyDraws = append(historyDraws, formatRecommendationSignature(draw.RedNumbers, draw.BlueNumbers))
	}

	return RecommendationBlocklist{
		HistoryDraws: historyDraws,
	}
}

func loadRecentRecommendationSignatures(userID uuid.UUID, code string, limit int) ([]string, error) {
	recommendations := make([]model.Recommendation, 0)
	if err := db.DB.Preload("Entries").
		Where("user_id = ? AND lottery_code = ?", userID, code).
		Order("created_at desc").
		Order("id desc").
		Limit(max(1, limit)).
		Find(&recommendations).Error; err != nil {
		return nil, err
	}

	signatures := make([]string, 0)
	seen := make(map[string]struct{})
	for _, recommendation := range recommendations {
		for _, entry := range recommendation.Entries {
			signature := formatRecommendationSignature(entry.RedNumbers, entry.BlueNumbers)
			if _, exists := seen[signature]; exists {
				continue
			}
			seen[signature] = struct{}{}
			signatures = append(signatures, signature)
		}
	}
	return signatures, nil
}
