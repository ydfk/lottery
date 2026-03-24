package lottery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type jisuResponse struct {
	Status int             `json:"status"`
	Msg    string          `json:"msg"`
	Result json.RawMessage `json:"result"`
}

type SyncResult struct {
	LotteryCode string `json:"lotteryCode"`
	Issue       string `json:"issue,omitempty"`
	SyncedCount int    `json:"syncedCount"`
}

type SyncOptions struct {
	Issue string
	Start int
	Count int
}

type BatchSyncResult struct {
	Results []SyncResult `json:"results"`
}

type saveDrawOptions struct {
	ExpectedIssue    string
	ExpectedDrawDate time.Time
}

func SyncLatestDraw(ctx context.Context, code string, issue string) (*SyncResult, error) {
	lotteryType, err := getLotteryType(code)
	if err != nil {
		return nil, err
	}
	definition, err := GetDefinition(code)
	if err != nil {
		return nil, err
	}
	if config.Current.Jisu.AppKey == "" {
		return nil, fmt.Errorf("未配置极速数据 appkey")
	}

	expectedIssue, expectedDrawDate, err := resolveLatestSyncTarget(definition, issue)
	if err != nil {
		return nil, err
	}

	item, err := fetchLatestDraw(ctx, lotteryType, issue)
	if err != nil {
		return nil, err
	}
	if issue != "" {
		itemIssue := normalizeIssueByCode(code, extractString(item, "issueno", "issue"))
		if itemIssue != "" && expectedIssue != "" && !containsString(issueAliases(code, expectedIssue), itemIssue) {
			return nil, fmt.Errorf("第三方返回的开奖期号与请求期号不一致: 请求 %s，返回 %s", expectedIssue, itemIssue)
		}
	}

	saved, savedIssue, err := saveDrawItem(lotteryType, item, saveDrawOptions{
		ExpectedIssue:    expectedIssue,
		ExpectedDrawDate: expectedDrawDate,
	})
	if err != nil {
		return nil, err
	}
	if err := settleByIssue(code, savedIssue); err != nil {
		return nil, err
	}

	result := &SyncResult{
		LotteryCode: code,
		Issue:       savedIssue,
	}
	if saved {
		result.SyncedCount = 1
	}
	return result, nil
}

func SyncDrawHistory(ctx context.Context, code string, options SyncOptions) (*SyncResult, error) {
	lotteryType, err := getLotteryType(code)
	if err != nil {
		return nil, err
	}
	if config.Current.Jisu.AppKey == "" {
		return nil, fmt.Errorf("未配置极速数据 appkey")
	}

	count := options.Count
	if count <= 0 {
		count = 100
	}

	syncedCount := 0
	offset := max(0, options.Start)
	remaining := count
	issues := make(map[string]struct{})
	for remaining > 0 {
		pageSize := min(20, remaining)
		items, err := fetchDrawHistory(ctx, lotteryType, options.Issue, offset, pageSize)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}

		for _, item := range items {
			saved, issue, saveErr := saveDrawItem(lotteryType, item, saveDrawOptions{})
			if saveErr != nil {
				return nil, saveErr
			}
			if issue != "" {
				issues[issue] = struct{}{}
			}
			if saved {
				syncedCount++
			}
		}

		if len(items) < pageSize {
			break
		}

		offset += len(items)
		remaining -= len(items)
	}

	for issue := range issues {
		if err := settleByIssue(code, issue); err != nil {
			return nil, err
		}
	}

	return &SyncResult{
		LotteryCode: code,
		SyncedCount: syncedCount,
	}, nil
}

func SyncMultipleDraws(ctx context.Context, codes []string, options SyncOptions) (*BatchSyncResult, error) {
	targetCodes := codes
	if len(targetCodes) == 0 {
		definitions := ListDefinitions()
		targetCodes = make([]string, 0, len(definitions))
		for _, definition := range definitions {
			if definition.Enabled {
				targetCodes = append(targetCodes, definition.Code)
			}
		}
	}

	results := make([]SyncResult, 0, len(targetCodes))
	for _, code := range targetCodes {
		result, err := SyncDrawHistory(ctx, code, options)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}

	return &BatchSyncResult{Results: results}, nil
}

func fetchLatestDraw(ctx context.Context, lotteryType model.LotteryType, issue string) (map[string]any, error) {
	requestURL := fmt.Sprintf(
		"%s/caipiao/query?appkey=%s&caipiaoid=%s&issueno=%s",
		strings.TrimRight(config.Current.Jisu.BaseURL, "/"),
		url.QueryEscape(config.Current.Jisu.AppKey),
		url.QueryEscape(lotteryType.RemoteLotteryID),
		url.QueryEscape(issue),
	)

	parsed, err := requestJisu(ctx, requestURL)
	if err != nil {
		return nil, err
	}
	return extractSingleItem(parsed.Result)
}

func fetchDrawHistory(ctx context.Context, lotteryType model.LotteryType, issue string, start int, count int) ([]map[string]any, error) {
	requestURL := fmt.Sprintf(
		"%s/caipiao/history?appkey=%s&caipiaoid=%s&issueno=%s&start=%d&num=%d",
		strings.TrimRight(config.Current.Jisu.BaseURL, "/"),
		url.QueryEscape(config.Current.Jisu.AppKey),
		url.QueryEscape(lotteryType.RemoteLotteryID),
		url.QueryEscape(issue),
		start,
		count,
	)

	parsed, err := requestJisu(ctx, requestURL)
	if err != nil {
		return nil, err
	}
	return extractItems(parsed.Result)
}

func requestJisu(ctx context.Context, requestURL string) (*jisuResponse, error) {
	startedAt := time.Now()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		logThirdPartyFailure("jisuapi", http.MethodGet, requestURL, map[string]any{
			"url": maskURL(requestURL),
		}, nil, 0, startedAt, err)
		return nil, err
	}

	client := &http.Client{Timeout: time.Duration(max(10, config.Current.Jisu.TimeoutSeconds)) * time.Second}
	response, err := client.Do(request)
	if err != nil {
		logThirdPartyFailure("jisuapi", http.MethodGet, requestURL, map[string]any{
			"url": maskURL(requestURL),
		}, nil, 0, startedAt, err)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logThirdPartyFailure("jisuapi", http.MethodGet, requestURL, map[string]any{
			"url": maskURL(requestURL),
		}, nil, response.StatusCode, startedAt, err)
		return nil, err
	}
	if response.StatusCode >= http.StatusBadRequest {
		requestErr := fmt.Errorf("开奖同步失败: %s", string(body))
		logThirdPartyFailure("jisuapi", http.MethodGet, requestURL, map[string]any{
			"url": maskURL(requestURL),
		}, body, response.StatusCode, startedAt, requestErr)
		return nil, requestErr
	}

	parsed := jisuResponse{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		logThirdPartyFailure("jisuapi", http.MethodGet, requestURL, map[string]any{
			"url": maskURL(requestURL),
		}, body, response.StatusCode, startedAt, err)
		return nil, err
	}
	if parsed.Status != 0 {
		requestErr := fmt.Errorf("开奖同步失败: %s", parsed.Msg)
		logThirdPartyFailure("jisuapi", http.MethodGet, requestURL, map[string]any{
			"url": maskURL(requestURL),
		}, body, response.StatusCode, startedAt, requestErr)
		return nil, requestErr
	}
	logThirdPartySuccess("jisuapi", http.MethodGet, requestURL, map[string]any{
		"url": maskURL(requestURL),
	}, body, response.StatusCode, startedAt)
	return &parsed, nil
}

func saveDrawItem(lotteryType model.LotteryType, item map[string]any, options saveDrawOptions) (bool, string, error) {
	definition, err := GetDefinition(lotteryType.Code)
	if err != nil {
		return false, "", err
	}

	issue := normalizeIssueByCode(lotteryType.Code, resolveValue(options.ExpectedIssue, extractString(item, "issueno", "issue")))
	if !isFinalDrawPayload(item) {
		if cleanupErr := cleanupUnfinalDrawByIssue(lotteryType.Code, issue); cleanupErr != nil {
			return false, issue, cleanupErr
		}
		return false, issue, nil
	}

	redNumbers, blueNumbers := parseDrawNumbers(lotteryType, item)
	if issue == "" || redNumbers == "" || blueNumbers == "" {
		return false, "", nil
	}

	draw := model.DrawResult{}
	err = db.DB.Where("lottery_code = ? AND issue = ?", lotteryType.Code, issue).First(&draw).Error
	isCreate := false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		draw = model.DrawResult{
			LotteryCode: lotteryType.Code,
			Issue:       issue,
		}
		isCreate = true
	} else if err != nil {
		return false, "", err
	}

	draw.DrawDate = resolveDrawDateForSave(definition, issue, options.ExpectedDrawDate, parseDrawDate(extractString(item, "opendate", "awardtime", "drawdate")))
	draw.RedNumbers = redNumbers
	draw.BlueNumbers = blueNumbers
	draw.SaleAmount = parseFloat(item["saleamount"])
	draw.PrizePoolAmount = parseFloatValues(item["poolamount"], item["totalmoney"])
	draw.Source = "jisuapi"
	draw.RawPayload = mustJSON(item)

	if isCreate {
		if err := db.DB.Create(&draw).Error; err != nil {
			return false, "", err
		}
	} else {
		if err := db.DB.Save(&draw).Error; err != nil {
			return false, "", err
		}
	}

	if err := saveDrawPrizes(draw.Id, item); err != nil {
		return false, "", err
	}

	return true, issue, nil
}

func parseDrawNumbers(lotteryType model.LotteryType, item map[string]any) (string, string) {
	mainNumbers := parseSpaceNumbers(extractString(item, "number", "awardnum"))
	referNumbers := parseSpaceNumbers(extractString(item, "refernumber", "blue"))

	if len(mainNumbers) >= lotteryType.RedCount && len(referNumbers) >= lotteryType.BlueCount {
		return formatNumbers(mainNumbers[:lotteryType.RedCount]), formatNumbers(referNumbers[:lotteryType.BlueCount])
	}

	combined := append(append([]int(nil), mainNumbers...), referNumbers...)
	if len(combined) < lotteryType.RedCount+lotteryType.BlueCount {
		return "", ""
	}

	redNumbers := combined[:lotteryType.RedCount]
	blueNumbers := combined[lotteryType.RedCount : lotteryType.RedCount+lotteryType.BlueCount]
	return formatNumbers(redNumbers), formatNumbers(blueNumbers)
}

func saveDrawPrizes(drawID uuid.UUID, item map[string]any) error {
	if err := db.DB.Where("draw_result_id = ?", drawID).Delete(&model.DrawPrize{}).Error; err != nil {
		return err
	}

	prizeItems := normalizePrizeItems(item["prize"])
	if len(prizeItems) == 0 {
		prizeItems = normalizePrizeItems(item)
	}
	if len(prizeItems) == 0 {
		return nil
	}

	records := make([]model.DrawPrize, 0, len(prizeItems))
	for _, prizeItem := range prizeItems {
		records = append(records, model.DrawPrize{
			DrawResultID: drawID,
			PrizeName:    normalizePrizeName(extractString(prizeItem, "name", "prizename")),
			PrizeRule:    extractString(prizeItem, "requirement", "require", "rule"),
			WinnerCount:  parseInt(prizeItem["num"]),
			SingleBonus:  parseFloat(prizeItem["singlebonus"]),
		})
	}

	if len(records) == 0 {
		return nil
	}
	return db.DB.Create(&records).Error
}

func normalizePrizeItems(raw any) []map[string]any {
	if raw == nil {
		return nil
	}

	switch actual := raw.(type) {
	case []any:
		result := make([]map[string]any, 0, len(actual))
		for _, item := range actual {
			if record, ok := item.(map[string]any); ok {
				result = append(result, record)
			}
		}
		return result
	case map[string]any:
		if _, hasName := actual["prizename"]; hasName {
			return []map[string]any{actual}
		}
		if _, hasName := actual["name"]; hasName {
			return []map[string]any{actual}
		}
	}

	return nil
}

func settleByIssue(code string, issue string) error {
	if issue == "" {
		return nil
	}
	if err := EvaluateTicketsByIssue(code, issue); err != nil {
		return err
	}
	return EvaluateRecommendationsByIssue(code, issue)
}

func cleanupUnfinalDrawByIssue(code string, issue string) error {
	if issue == "" {
		return nil
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		draws := make([]model.DrawResult, 0)
		if err := tx.Where("lottery_code = ? AND issue IN ?", code, issueAliases(code, issue)).Find(&draws).Error; err != nil {
			return err
		}
		drawIDs := make([]uuid.UUID, 0, len(draws))
		for _, draw := range draws {
			if !isUnfinalDrawResult(draw) {
				continue
			}
			drawIDs = append(drawIDs, draw.Id)
		}
		if len(drawIDs) == 0 {
			return nil
		}
		if err := tx.Where("draw_result_id IN ?", drawIDs).Delete(&model.DrawPrize{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", drawIDs).Delete(&model.DrawResult{}).Error; err != nil {
			return err
		}
		if err := resetTicketsPendingByIssueWithDB(tx, code, issue); err != nil {
			return err
		}
		return resetRecommendationsPendingByIssueWithDB(tx, code, issue)
	})
}

func isUnfinalDrawResult(draw model.DrawResult) bool {
	return strings.Contains(draw.RawPayload, `"prize":false`)
}

func normalizePrizeName(value string) string {
	switch value {
	case "1", "一等奖":
		return "一等奖"
	case "2", "二等奖":
		return "二等奖"
	case "3", "三等奖":
		return "三等奖"
	case "4", "四等奖":
		return "四等奖"
	case "5", "五等奖":
		return "五等奖"
	case "6", "六等奖":
		return "六等奖"
	default:
		return value
	}
}

func parseSpaceNumbers(value string) []int {
	tokens := numberPattern.FindAllString(value, -1)
	numbers := make([]int, 0, len(tokens))
	for _, token := range tokens {
		number, err := strconv.Atoi(token)
		if err == nil {
			numbers = append(numbers, number)
		}
	}
	return numbers
}

func parseDrawDate(value string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		timestamp, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return timestamp
		}
	}
	return time.Time{}
}

func resolveLatestSyncTarget(definition Definition, requestedIssue string) (string, time.Time, error) {
	canonicalIssue := normalizeIssueByCode(definition.Code, requestedIssue)
	if canonicalIssue != "" {
		drawDate, ok, err := resolveLocalDrawByIssue(definition, canonicalIssue)
		if err != nil {
			return "", time.Time{}, err
		}
		if ok {
			return canonicalIssue, drawDate, nil
		}
		return canonicalIssue, time.Time{}, nil
	}

	issue, drawDate, ok, err := resolveLatestLocalDraw(definition, time.Now())
	if err != nil {
		return "", time.Time{}, err
	}
	if ok {
		return issue, drawDate, nil
	}
	return "", time.Time{}, nil
}

func resolveDrawDateForSave(definition Definition, issue string, expectedDrawDate time.Time, providerDrawDate time.Time) time.Time {
	if !expectedDrawDate.IsZero() {
		return expectedDrawDate
	}

	if issue != "" {
		localDrawDate, ok, err := resolveLocalDrawByIssue(definition, issue)
		if err == nil && ok {
			return localDrawDate
		}
	}

	if !providerDrawDate.IsZero() {
		return providerDrawDate
	}
	return time.Time{}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
