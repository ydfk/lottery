package lottery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TicketUploadStatusUploaded   = "uploaded"
	TicketUploadStatusRecognized = "recognized"
	TicketUploadStatusFailed     = "failed"
	TicketUploadStatusSaving     = "saving"
	TicketUploadStatusSaved      = "saved"
)

var ErrDuplicateTicket = errors.New("相同票据已存在，不能重复录入")

type UploadTicketImageInput struct {
	UserID           string
	Code             string
	ImagePath        string
	OriginalFilename string
}

type RecognizeUploadedTicketInput struct {
	UserID   string
	Code     string
	UploadID string
	OCRText  string
}

type CreateTicketInput struct {
	UserID           string
	Code             string
	UploadID         string
	RecommendationID string
	Issue            string
	DrawDate         time.Time
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
	userUUID, err := parseRequiredUserID(input.UserID)
	if err != nil {
		return nil, err
	}

	upload := model.TicketUpload{
		UserID:           &userUUID,
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
	upload, err := getTicketUpload(input.UserID, input.Code, input.UploadID)
	if err != nil {
		return nil, err
	}
	if upload.Status == TicketUploadStatusSaved {
		return nil, fmt.Errorf("这张图片已经入库，请重新上传新图片")
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
	if input.UploadID == "" {
		return nil, fmt.Errorf("请先上传彩票图片")
	}

	recommendation, recommendationID, err := resolveRecommendation(input.UserID, input.Code, input.RecommendationID)
	if err != nil {
		return nil, err
	}

	purchasedAt := input.PurchasedAt
	if purchasedAt.IsZero() {
		purchasedAt = time.Now()
	}

	ticketID := ""
	code := ""
	issue := ""
	shouldEvaluate := false

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		upload, reserveErr := reserveTicketUpload(tx, input.UserID, input.Code, input.UploadID)
		if reserveErr != nil {
			return reserveErr
		}

		code, reserveErr = resolveCreateTicketCode(input.Code, upload.LotteryCode, recommendation)
		if reserveErr != nil {
			return reserveErr
		}

		definition, definitionErr := GetDefinition(code)
		if definitionErr != nil {
			return definitionErr
		}

		entries := input.Entries
		if len(entries) == 0 {
			if upload.RecognitionPayload == "" {
				return fmt.Errorf("请先识别号码，或手动填写完整号码后保存")
			}
			entries, reserveErr = getRecognizedEntries(upload)
			if reserveErr != nil {
				return reserveErr
			}
		}

		if len(entries) > 0 {
			entries, reserveErr = normalizeParsedEntries(definition, entries)
			if reserveErr != nil {
				return reserveErr
			}
		}

		issue = input.Issue
		if issue == "" {
			issue = upload.RecognitionIssue
		}
		if issue == "" && recommendation != nil {
			issue = recommendation.Issue
		}
		if issue == "" {
			return fmt.Errorf("未识别到期号，请手动补充期号后重试")
		}
		issue = normalizeIssueByCode(code, issue)

		drawDate := input.DrawDate

		if reserveErr = validateDuplicateTicket(tx, input.UserID, code, issue, entries); reserveErr != nil {
			return reserveErr
		}

		ticket, createErr := createTicketRecord(tx, input.UserID, code, recommendationID, issue, drawDate, "upload", upload.ImagePath, upload.RecognizedText, purchasedAt, input.CostAmount, input.Notes, entries)
		if createErr != nil {
			if isUniqueConstraintError(createErr) {
				return ErrDuplicateTicket
			}
			return createErr
		}

		upload.LotteryCode = code
		upload.Status = TicketUploadStatusSaved
		if err := tx.Model(&model.TicketUpload{}).
			Where("id = ? AND status = ?", upload.Id, TicketUploadStatusSaving).
			Updates(map[string]any{
				"lottery_code": upload.LotteryCode,
				"status":       upload.Status,
			}).Error; err != nil {
			return err
		}

		ticketID = ticket.Id.String()
		shouldEvaluate = len(entries) > 0
		input.DrawDate = drawDate
		return nil
	}); err != nil {
		return nil, err
	}

	if shouldDeferSettlement(code, &input.DrawDate) {
		return GetTicketDetail(ticketID, input.UserID)
	}
	ensureIssueDrawSynced(ctx, code, issue)
	if shouldEvaluate {
		if err := EvaluateTicket(ticketID); err != nil {
			logger.Warn("票据自动判奖失败 %s/%s: %v", code, ticketID, err)
		}
	}
	return GetTicketDetail(ticketID, input.UserID)
}

func createTicketRecord(tx *gorm.DB, userID string, code string, recommendationID *uuid.UUID, issue string, drawDate time.Time, source string, imagePath string, recognizedText string, purchasedAt time.Time, costAmount float64, notes string, entries []ParsedEntry) (*model.Ticket, error) {
	userUUID, err := parseRequiredUserID(userID)
	if err != nil {
		return nil, err
	}

	totalCost := costAmount
	if totalCost <= 0 {
		totalCost = calculateEntriesCost(entries)
	}

	var manualDrawDate *time.Time
	if !drawDate.IsZero() {
		value := drawDate
		manualDrawDate = &value
	}

	var entrySignature *string
	if len(entries) > 0 {
		signature := buildTicketEntriesHash(entries)
		entrySignature = &signature
	}

	ticket := model.Ticket{
		UserID:           &userUUID,
		LotteryCode:      code,
		RecommendationID: recommendationID,
		Issue:            issue,
		EntrySignature:   entrySignature,
		ManualDrawDate:   manualDrawDate,
		Source:           resolveValue(source, "upload"),
		ImagePath:        imagePath,
		RecognizedText:   recognizedText,
		Status:           TicketStatusPending,
		CostAmount:       totalCost,
		PurchasedAt:      purchasedAt,
		Notes:            notes,
	}
	if err := tx.Create(&ticket).Error; err != nil {
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
		if err := tx.Create(&records).Error; err != nil {
			return nil, err
		}
	}

	return &ticket, nil
}

func validateDuplicateTicket(tx *gorm.DB, userID string, code string, issue string, entries []ParsedEntry) error {
	if len(entries) == 0 {
		return nil
	}

	signature := buildTicketEntriesHash(entries)
	var count int64
	if err := currentUserScope(tx.Model(&model.Ticket{}), userID).
		Where("lottery_code = ? AND issue IN ? AND entry_signature = ?", code, issueAliases(code, issue), signature).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateTicket
	}
	return nil
}

func reserveTicketUpload(tx *gorm.DB, userID string, code string, uploadID string) (model.TicketUpload, error) {
	upload, err := getTicketUploadWithDB(tx, userID, code, uploadID)
	if err != nil {
		return upload, err
	}
	if upload.Status == TicketUploadStatusSaved {
		return upload, fmt.Errorf("这张图片已经入库，请勿重复保存")
	}
	if upload.Status == TicketUploadStatusSaving {
		return upload, fmt.Errorf("这张图片正在处理中，请稍后重试")
	}

	result := currentUserScope(tx.Model(&model.TicketUpload{}), userID).
		Where("id = ? AND status = ?", uploadID, upload.Status).
		Update("status", TicketUploadStatusSaving)
	if result.Error != nil {
		return upload, result.Error
	}
	if result.RowsAffected == 0 {
		return upload, fmt.Errorf("这张图片正在处理中，请稍后重试")
	}

	upload.Status = TicketUploadStatusSaving
	return upload, nil
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

func buildTicketEntriesHash(entries []ParsedEntry) string {
	sum := sha256.Sum256([]byte(buildTicketEntriesSignature(entries)))
	return hex.EncodeToString(sum[:])
}

func getTicketUpload(userID string, code string, uploadID string) (model.TicketUpload, error) {
	return getTicketUploadWithDB(db.DB, userID, code, uploadID)
}

func getTicketUploadWithDB(database *gorm.DB, userID string, code string, uploadID string) (model.TicketUpload, error) {
	upload := model.TicketUpload{}
	if err := currentUserScope(database, userID).First(&upload, "id = ?", uploadID).Error; err != nil {
		return upload, err
	}
	if code != "" && upload.LotteryCode != "" && upload.LotteryCode != code {
		return upload, fmt.Errorf("上传记录的彩票类型与当前录入类型不一致")
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

func resolveRecommendation(userID string, code string, recommendationID string) (*model.Recommendation, *uuid.UUID, error) {
	if recommendationID == "" {
		return nil, nil, nil
	}

	parsedID, err := uuid.Parse(recommendationID)
	if err != nil {
		return nil, nil, fmt.Errorf("推荐记录不存在")
	}

	recommendation := model.Recommendation{}
	query := currentUserScope(db.DB, userID)
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

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "UNIQUE constraint failed") || strings.Contains(message, "duplicate key value")
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
