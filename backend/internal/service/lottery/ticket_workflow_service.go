package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"
)

const (
	TicketUploadStatusUploaded   = "uploaded"
	TicketUploadStatusRecognized = "recognized"
	TicketUploadStatusFailed     = "failed"
	TicketUploadStatusSaved      = "saved"
)

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
	Code        string
	UploadID    string
	Issue       string
	PurchasedAt time.Time
	Notes       string
	Entries     []ParsedEntry
}

type TicketUploadDetail struct {
	model.TicketUpload
	ImageURL string `json:"imageUrl"`
}

type TicketRecognitionDraft struct {
	Upload     TicketUploadDetail `json:"upload"`
	Issue      string             `json:"issue"`
	RawText    string             `json:"rawText"`
	Confidence float64            `json:"confidence"`
	Entries    []ParsedEntry      `json:"entries"`
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

	lotteryType, err := getLotteryType(input.Code)
	if err != nil {
		return nil, err
	}
	definition, err := GetDefinition(input.Code)
	if err != nil {
		return nil, err
	}

	recognized, err := recognizeTicket(ctx, definition, lotteryType, upload.ImagePath, input.OCRText)
	if err != nil {
		upload.Status = TicketUploadStatusFailed
		upload.ErrorMessage = err.Error()
		if saveErr := db.DB.Save(&upload).Error; saveErr != nil {
			return nil, saveErr
		}
		return nil, err
	}

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
		Upload:     *buildTicketUploadDetail(upload),
		Issue:      recognized.Issue,
		RawText:    recognized.RawText,
		Confidence: recognized.Confidence,
		Entries:    recognized.Entries,
	}, nil
}

func CreateTicket(ctx context.Context, input CreateTicketInput) (*TicketDetail, error) {
	upload, err := getTicketUpload(input.Code, input.UploadID)
	if err != nil {
		return nil, err
	}

	definition, err := GetDefinition(input.Code)
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
	if issue == "" {
		return nil, fmt.Errorf("未识别到期号，请手动补充期号后重试")
	}

	purchasedAt := input.PurchasedAt
	if purchasedAt.IsZero() {
		purchasedAt = time.Now()
	}

	ticket, err := createTicketRecord(input.Code, issue, upload.ImagePath, upload.RecognizedText, purchasedAt, input.Notes, entries)
	if err != nil {
		return nil, err
	}

	upload.Status = TicketUploadStatusSaved
	if err := db.DB.Save(&upload).Error; err != nil {
		return nil, err
	}

	if err := EvaluateTicket(ticket.Id.String()); err != nil {
		return nil, err
	}
	return GetTicketDetail(ticket.Id.String())
}

func createTicketRecord(code string, issue string, imagePath string, recognizedText string, purchasedAt time.Time, notes string, entries []ParsedEntry) (*model.Ticket, error) {
	ticket := model.Ticket{
		LotteryCode:    code,
		Issue:          issue,
		Source:         "upload",
		ImagePath:      imagePath,
		RecognizedText: recognizedText,
		Status:         TicketStatusPending,
		CostAmount:     float64(len(entries) * 2),
		PurchasedAt:    purchasedAt,
		Notes:          notes,
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

func getTicketUpload(code string, uploadID string) (model.TicketUpload, error) {
	upload := model.TicketUpload{}
	if err := db.DB.First(&upload, "id = ? AND lottery_code = ?", uploadID, code).Error; err != nil {
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
			Red:  parseCSVNumbers(formatNumbers(entry.Red)),
			Blue: parseCSVNumbers(formatNumbers(entry.Blue)),
		})
	}
	return result, nil
}

func validateParsedEntry(definition Definition, entry ParsedEntry) error {
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
	return nil
}
