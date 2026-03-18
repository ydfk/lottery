package lottery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
)

const (
	TicketUploadStatusUploaded   = "uploaded"
	TicketUploadStatusRecognized = "recognized"
	TicketUploadStatusFailed     = "failed"
	TicketUploadStatusSaved      = "saved"
)

var ErrDuplicateTicket = errors.New("相同票据已存在，不能重复录入")

type UploadTicketImageInput struct {
	Code             string
	ImagePath        string
	OriginalFilename string
}

type RecognizeUploadedTicketInput struct {
	Code     string
	UploadID string
	OCRText  string
}

type CreateTicketInput struct {
	Code             string
	UploadID         string
	RecommendationID string
	Issue            string
	PurchasedAt      time.Time
	CostAmount       float64
	Notes            string
	Entries          []ParsedEntry
}

type TicketUploadDetail struct {
	model.TicketUpload
	ImageURL string `json:"imageUrl"`
}

type TicketRecognitionDraft struct {
	Upload      TicketUploadDetail `json:"upload"`
	LotteryCode string             `json:"lotteryCode"`
	Issue       string             `json:"issue"`
	DrawDate    string             `json:"drawDate"`
	CostAmount  float64            `json:"costAmount"`
	RawText     string             `json:"rawText"`
	Confidence  float64            `json:"confidence"`
	Entries     []ParsedEntry      `json:"entries"`
}

func UploadTicketImage(input UploadTicketImageInput) (*TicketUploadDetail, error) {
	upload := model.TicketUpload{
		LotteryCode:      input.Code,
		Status:           TicketUploadStatusUploaded,
		OriginalFilename: input.OriginalFilename,
		ImagePath:        input.ImagePath,
	}
	if err := db.DB.Create(&upload).Error; err != nil {
		return nil, err
	}
	return buildTicketUploadDetail(upload), nil
}

func RecognizeUploadedTicket(ctx context.Context, input RecognizeUploadedTicketInput) (*TicketRecognitionDraft, error) {
	upload, err := getTicketUpload(input.Code, input.UploadID)
	if err != nil {
		return nil, err
	}

	recognized, err := recognizeTicket(ctx, resolveTicketCode(input.Code, upload.LotteryCode), upload.ImagePath, input.OCRText)
	if err != nil {
		upload.Status = TicketUploadStatusFailed
		upload.ErrorMessage = err.Error()
		if saveErr := db.DB.Save(&upload).Error; saveErr != nil {
			return nil, saveErr
		}
		return nil, err
	}

	upload.LotteryCode = recognized.LotteryCode
	upload.Status = TicketUploadStatusRecognized
	upload.ErrorMessage = ""
	upload.RecognizedText = recognized.RawText
	upload.RecognitionIssue = recognized.Issue
	upload.RecognitionConfidence = recognized.Confidence
	upload.RecognitionPayload = mustJSON(recognized)
	if err := db.DB.Save(&upload).Error; err != nil {
		return nil, err
	}

	return &TicketRecognitionDraft{
		Upload:      *buildTicketUploadDetail(upload),
		LotteryCode: recognized.LotteryCode,
		Issue:       recognized.Issue,
		DrawDate:    recognized.DrawDate,
		CostAmount:  recognized.CostAmount,
		RawText:     recognized.RawText,
		Confidence:  recognized.Confidence,
		Entries:     recognized.Entries,
	}, nil
}

func CreateTicket(ctx context.Context, input CreateTicketInput) (*TicketDetail, error) {
	upload, err := getTicketUpload(input.Code, input.UploadID)
	if err != nil {
		return nil, err
	}

	recommendation, recommendationID, err := resolveRecommendation(input.Code, input.RecommendationID)
	if err != nil {
		return nil, err
	}

	code, err := resolveCreateTicketCode(input.Code, upload.LotteryCode, recommendation)
	if err != nil {
		return nil, err
	}

	definition, err := GetDefinition(code)
	if err != nil {
		return nil, err
	}

	entries := input.Entries
	if len(entries) == 0 {
		entries, err = getRecognizedEntries(upload)
		if err != nil {
			return nil, err
		}
	}

	entries, err = normalizeParsedEntries(definition, entries)
	if err != nil {
		return nil, err
	}

	issue := input.Issue
	if issue == "" {
		issue = upload.RecognitionIssue
	}
	if issue == "" && recommendation != nil {
		issue = recommendation.Issue
	}
	if issue == "" {
		return nil, fmt.Errorf("未识别到期号，请手动补充期号后重试")
	}
	issue = normalizeIssueByCode(code, issue)

	purchasedAt := input.PurchasedAt
	if purchasedAt.IsZero() {
		purchasedAt = time.Now()
	}

	if err := validateDuplicateTicket(code, issue, input.CostAmount, entries); err != nil {
		return nil, err
	}

	ticket, err := createTicketRecord(code, recommendationID, issue, upload.ImagePath, upload.RecognizedText, purchasedAt, input.CostAmount, input.Notes, entries)
	if err != nil {
		return nil, err
	}

	upload.LotteryCode = code
	upload.Status = TicketUploadStatusSaved
	if err := db.DB.Save(&upload).Error; err != nil {
		return nil, err
	}

	ensureIssueDrawSynced(ctx, code, issue)

	if err := EvaluateTicket(ticket.Id.String()); err != nil {
		return nil, err
	}
	return GetTicketDetail(ticket.Id.String())
}

func createTicketRecord(code string, recommendationID *uuid.UUID, issue string, imagePath string, recognizedText string, purchasedAt time.Time, costAmount float64, notes string, entries []ParsedEntry) (*model.Ticket, error) {
	totalCost := costAmount
	if totalCost <= 0 {
		totalCost = calculateEntriesCost(entries)
	}

	ticket := model.Ticket{
		LotteryCode:      code,
		RecommendationID: recommendationID,
		Issue:            issue,
		Source:           "upload",
		ImagePath:        imagePath,
		RecognizedText:   recognizedText,
		Status:           TicketStatusPending,
		CostAmount:       totalCost,
		PurchasedAt:      purchasedAt,
		Notes:            notes,
	}
	if err := db.DB.Create(&ticket).Error; err != nil {
		return nil, err
	}

	records := make([]model.TicketEntry, 0, len(entries))
	for index, item := range entries {
		records = append(records, model.TicketEntry{
			TicketID:     ticket.Id,
			Sequence:     index + 1,
			RedNumbers:   formatNumbers(item.Red),
			BlueNumbers:  formatNumbers(item.Blue),
			Multiple:     resolveEntryMultiple(item),
			IsAdditional: item.IsAdditional,
			MatchSummary: "待开奖",
		})
	}
	if len(records) > 0 {
		if err := db.DB.Create(&records).Error; err != nil {
			return nil, err
		}
	}

	return &ticket, nil
}

func validateDuplicateTicket(code string, issue string, costAmount float64, entries []ParsedEntry) error {
	candidates := make([]model.Ticket, 0)
	if err := db.DB.Preload("Entries").Where("lottery_code = ? AND issue IN ?", code, issueAliases(code, issue)).Find(&candidates).Error; err != nil {
		return err
	}

	targetCost := costAmount
	if targetCost <= 0 {
		targetCost = calculateEntriesCost(entries)
	}
	targetSignature := buildTicketEntriesSignature(entries)
	for _, candidate := range candidates {
		if len(candidate.Entries) != len(entries) {
			continue
		}
		if targetCost > 0 && candidate.CostAmount > 0 && !isSameAmount(candidate.CostAmount, targetCost) {
			continue
		}
		if buildStoredTicketEntriesSignature(candidate.Entries) == targetSignature {
			return ErrDuplicateTicket
		}
	}

	return nil
}

func buildTicketEntriesSignature(entries []ParsedEntry) string {
	signature := ""
	for index, entry := range entries {
		if index > 0 {
			signature += "|"
		}
		signature += fmt.Sprintf(
			"%d:%s:%s:%d:%t",
			index+1,
			formatNumbers(entry.Red),
			formatNumbers(entry.Blue),
			resolveEntryMultiple(entry),
			entry.IsAdditional,
		)
	}
	return signature
}

func buildStoredTicketEntriesSignature(entries []model.TicketEntry) string {
	sort.Slice(entries, func(left int, right int) bool {
		if entries[left].Sequence == entries[right].Sequence {
			return entries[left].CreatedAt.Before(entries[right].CreatedAt)
		}
		return entries[left].Sequence < entries[right].Sequence
	})

	signature := ""
	for index, entry := range entries {
		if index > 0 {
			signature += "|"
		}
		signature += fmt.Sprintf(
			"%d:%s:%s:%d:%t",
			entry.Sequence,
			entry.RedNumbers,
			entry.BlueNumbers,
			max(1, entry.Multiple),
			entry.IsAdditional,
		)
	}
	return signature
}

func isSameAmount(left float64, right float64) bool {
	delta := left - right
	if delta < 0 {
		delta = -delta
	}
	return delta < 0.0001
}

func getTicketUpload(code string, uploadID string) (model.TicketUpload, error) {
	upload := model.TicketUpload{}
	query := db.DB
	if code != "" {
		query = query.Where("lottery_code = ?", code)
	}
	if err := query.First(&upload, "id = ?", uploadID).Error; err != nil {
		return upload, err
	}
	return upload, nil
}

func buildTicketUploadDetail(upload model.TicketUpload) *TicketUploadDetail {
	return &TicketUploadDetail{
		TicketUpload: upload,
		ImageURL:     buildPublicImageURL(upload.ImagePath),
	}
}

func getRecognizedEntries(upload model.TicketUpload) ([]ParsedEntry, error) {
	if upload.RecognitionPayload == "" {
		return nil, fmt.Errorf("请先调用识别接口获取号码结果")
	}

	recognized := RecognitionResult{}
	if err := json.Unmarshal([]byte(upload.RecognitionPayload), &recognized); err != nil {
		return nil, fmt.Errorf("识别结果无法解析，请重新识别")
	}
	if len(recognized.Entries) == 0 {
		return nil, fmt.Errorf("识别结果中没有可入库的号码")
	}
	return recognized.Entries, nil
}

func resolveRecommendation(code string, recommendationID string) (*model.Recommendation, *uuid.UUID, error) {
	if recommendationID == "" {
		return nil, nil, nil
	}

	parsedID, err := uuid.Parse(recommendationID)
	if err != nil {
		return nil, nil, fmt.Errorf("推荐记录不存在")
	}

	recommendation := model.Recommendation{}
	query := db.DB
	if code != "" {
		query = query.Where("lottery_code = ?", code)
	}
	if err := query.First(&recommendation, "id = ?", parsedID).Error; err != nil {
		return nil, nil, fmt.Errorf("推荐记录不存在")
	}
	return &recommendation, &parsedID, nil
}

func resolveCreateTicketCode(code string, uploadCode string, recommendation *model.Recommendation) (string, error) {
	if code != "" {
		return code, nil
	}
	if uploadCode != "" {
		return uploadCode, nil
	}
	if recommendation != nil {
		return recommendation.LotteryCode, nil
	}
	return "", fmt.Errorf("未识别出彩票类型，请先完成识别或手动指定彩种")
}

func resolveTicketCode(primary string, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func ensureIssueDrawSynced(ctx context.Context, code string, issue string) {
	if issue == "" {
		return
	}

	var count int64
	if err := db.DB.Model(&model.DrawResult{}).Where("lottery_code = ? AND issue IN ?", code, issueAliases(code, issue)).Count(&count).Error; err == nil && count > 0 {
		return
	}

	if _, err := SyncLatestDraw(ctx, code, issue); err != nil {
		return
	}
}

func normalizeParsedEntries(definition Definition, entries []ParsedEntry) ([]ParsedEntry, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("至少需要一注号码")
	}

	result := make([]ParsedEntry, 0, len(entries))
	for _, entry := range entries {
		if err := validateParsedEntry(definition, entry); err != nil {
			return nil, err
		}
		result = append(result, ParsedEntry{
			Red:          parseCSVNumbers(formatNumbers(entry.Red)),
			Blue:         parseCSVNumbers(formatNumbers(entry.Blue)),
			Multiple:     resolveEntryMultiple(entry),
			IsAdditional: entry.IsAdditional,
		})
	}
	return result, nil
}

func validateParsedEntry(definition Definition, entry ParsedEntry) error {
	if definition.Code == "ssq" && entry.IsAdditional {
		return fmt.Errorf("双色球不支持追加")
	}
	if len(entry.Red) != definition.RedCount {
		return fmt.Errorf("红球数量不正确，应为 %d 个", definition.RedCount)
	}
	if len(entry.Blue) != definition.BlueCount {
		return fmt.Errorf("蓝球数量不正确，应为 %d 个", definition.BlueCount)
	}
	if containsDuplicate(entry.Red) {
		return fmt.Errorf("红球号码不能重复")
	}
	if containsDuplicate(entry.Blue) {
		return fmt.Errorf("蓝球号码不能重复")
	}

	for _, value := range entry.Red {
		if value < definition.RedMin || value > definition.RedMax {
			return fmt.Errorf("红球号码超出范围，应在 %d-%d 之间", definition.RedMin, definition.RedMax)
		}
	}
	for _, value := range entry.Blue {
		if value < definition.BlueMin || value > definition.BlueMax {
			return fmt.Errorf("蓝球号码超出范围，应在 %d-%d 之间", definition.BlueMin, definition.BlueMax)
		}
	}
	if resolveEntryMultiple(entry) <= 0 {
		return fmt.Errorf("注数/倍数必须大于 0")
	}
	return nil
}

func resolveEntryMultiple(entry ParsedEntry) int {
	if entry.Multiple <= 0 {
		return 1
	}
	return entry.Multiple
}
