package lottery

import (
	"context"
	"fmt"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
)

type ScanTicketInput struct {
	Code        string
	Issue       string
	ImagePath   string
	OCRText     string
	PurchasedAt time.Time
	Notes       string
}

type TicketDetail struct {
	model.Ticket
	ImageURL        string     `json:"imageUrl"`
	DrawDate        *time.Time `json:"drawDate"`
	DrawRedNumbers  string     `json:"drawRedNumbers"`
	DrawBlueNumbers string     `json:"drawBlueNumbers"`
}

func ScanTicket(ctx context.Context, input ScanTicketInput) (*TicketDetail, error) {
	upload, err := UploadTicketImage(UploadTicketImageInput{
		Code:             input.Code,
		ImagePath:        input.ImagePath,
		OriginalFilename: "",
	})
	if err != nil {
		return nil, err
	}

	recognized, err := RecognizeUploadedTicket(ctx, RecognizeUploadedTicketInput{
		Code:     input.Code,
		UploadID: upload.Id.String(),
		OCRText:  input.OCRText,
	})
	if err != nil {
		return nil, err
	}

	return CreateTicket(ctx, CreateTicketInput{
		Code:        input.Code,
		UploadID:    upload.Id.String(),
		Issue:       resolveValue(input.Issue, recognized.Issue),
		PurchasedAt: input.PurchasedAt,
		Notes:       input.Notes,
	})
}

func recognizeTicket(ctx context.Context, code string, imagePath string, ocrText string) (*RecognitionResult, error) {
	if ocrText != "" {
		if code != "" {
			return ParseLotteryText(code, ocrText)
		}
		return DetectLotteryByText(ocrText)
	}

	var lotteryType *model.LotteryType
	provider := config.Current.Vision.Provider
	if code != "" {
		typeRecord, err := getLotteryType(code)
		if err != nil {
			return nil, err
		}
		lotteryType = &typeRecord
		provider = resolveValue(config.Current.Vision.Provider, typeRecord.VisionProvider)
	}

	recognizer := newVisionRecognizer(provider)
	if recognizer == nil {
		return nil, fmt.Errorf("未配置视觉模型，请填写 OCR 文本作为降级输入")
	}

	recognized, err := recognizer.Recognize(ctx, lotteryType, imagePath)
	if err != nil {
		return nil, err
	}
	if len(recognized.Entries) == 0 && recognized.RawText != "" {
		if code != "" {
			return ParseLotteryText(code, recognized.RawText)
		}
		return DetectLotteryByText(recognized.RawText)
	}
	if len(recognized.Entries) == 0 {
		return nil, fmt.Errorf("未从图片中识别到有效号码")
	}
	if recognized.LotteryCode == "" && code != "" {
		recognized.LotteryCode = code
	}
	return recognized, nil
}

func EvaluatePendingTickets(code string) error {
	tickets := make([]model.Ticket, 0)
	if err := db.DB.Where("lottery_code = ? AND status = ?", code, TicketStatusPending).Find(&tickets).Error; err != nil {
		return err
	}
	for _, ticket := range tickets {
		if err := EvaluateTicket(ticket.Id.String()); err != nil {
			return err
		}
	}
	return nil
}

func EvaluateTicketsByIssue(code string, issue string) error {
	tickets := make([]model.Ticket, 0)
	if err := db.DB.Where("lottery_code = ? AND issue IN ?", code, issueAliases(code, issue)).Find(&tickets).Error; err != nil {
		return err
	}
	for _, ticket := range tickets {
		if err := EvaluateTicket(ticket.Id.String()); err != nil {
			return err
		}
	}
	return nil
}

func EvaluateTicket(ticketID string) error {
	ticket := model.Ticket{}
	if err := db.DB.Preload("Entries").First(&ticket, "id = ?", ticketID).Error; err != nil {
		return err
	}
	canonicalIssue := normalizeIssueByCode(ticket.LotteryCode, ticket.Issue)
	if canonicalIssue != "" && canonicalIssue != ticket.Issue {
		ticket.Issue = canonicalIssue
		if err := db.DB.Omit("Entries").Save(&ticket).Error; err != nil {
			return err
		}
	}

	draw := model.DrawResult{}
	if err := db.DB.Preload("PrizeDetails").Where("lottery_code = ? AND issue IN ?", ticket.LotteryCode, issueAliases(ticket.LotteryCode, ticket.Issue)).First(&draw).Error; err != nil {
		return nil
	}

	prizeMap := make(map[string]float64, len(draw.PrizeDetails))
	for _, prize := range draw.PrizeDetails {
		prizeMap[normalizePrizeName(prize.PrizeName)] = prize.SingleBonus
	}

	totalPrize := 0.0
	hasWinning := false
	for _, entry := range ticket.Entries {
		result := JudgeNumbers(ticket.LotteryCode, entry.RedNumbers, entry.BlueNumbers, entry.IsAdditional, draw, prizeMap)
		entry.IsWinning = result.IsWinning
		entry.PrizeName = result.PrizeName
		entry.PrizeAmount = result.PrizeAmount * float64(max(1, entry.Multiple))
		entry.MatchSummary = result.MatchSummary
		totalPrize += entry.PrizeAmount
		hasWinning = hasWinning || result.IsWinning
		if err := db.DB.Save(&entry).Error; err != nil {
			return err
		}
	}

	checkedAt := time.Now()
	ticket.CheckedAt = &checkedAt
	ticket.PrizeAmount = totalPrize
	if hasWinning {
		ticket.Status = TicketStatusWon
	} else {
		ticket.Status = TicketStatusNotWon
	}
	return db.DB.Omit("Entries").Save(&ticket).Error
}

func RecheckTicket(ctx context.Context, ticketID string, code string) (*TicketDetail, error) {
	ticket := model.Ticket{}
	query := db.DB.Preload("Entries")
	if code != "" {
		query = query.Where("lottery_code = ?", code)
	}
	if err := query.First(&ticket, "id = ?", ticketID).Error; err != nil {
		return nil, err
	}

	ticketIssue := normalizeIssueByCode(ticket.LotteryCode, ticket.Issue)
	if ticketIssue == "" {
		return nil, fmt.Errorf("票据期号不能为空")
	}
	if _, err := SyncLatestDraw(ctx, ticket.LotteryCode, ticketIssue); err != nil {
		return nil, err
	}
	if err := EvaluateTicket(ticket.Id.String()); err != nil {
		return nil, err
	}
	return GetTicketDetail(ticket.Id.String())
}

func GetTicketDetail(ticketID string) (*TicketDetail, error) {
	ticket := model.Ticket{}
	if err := db.DB.Preload("Entries").First(&ticket, "id = ?", ticketID).Error; err != nil {
		return nil, err
	}
	return buildTicketDetail(ticket), nil
}

func ListTickets(code string, limit int) ([]TicketDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	tickets := make([]model.Ticket, 0)
	if err := db.DB.Preload("Entries").Where("lottery_code = ?", code).Order("created_at desc").Limit(limit).Find(&tickets).Error; err != nil {
		return nil, err
	}

	result := make([]TicketDetail, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, *buildTicketDetail(ticket))
	}
	return result, nil
}

func ListAllTickets(limit int) ([]TicketDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	tickets := make([]model.Ticket, 0)
	if err := db.DB.Preload("Entries").Order("created_at desc").Limit(limit).Find(&tickets).Error; err != nil {
		return nil, err
	}

	result := make([]TicketDetail, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, *buildTicketDetail(ticket))
	}
	return result, nil
}

func buildTicketDetail(ticket model.Ticket) *TicketDetail {
	detail := &TicketDetail{
		Ticket:   ticket,
		ImageURL: buildPublicImageURL(ticket.ImagePath),
	}

	draw := model.DrawResult{}
	if err := db.DB.Select("draw_date", "red_numbers", "blue_numbers").Where("lottery_code = ? AND issue IN ?", ticket.LotteryCode, issueAliases(ticket.LotteryCode, ticket.Issue)).First(&draw).Error; err == nil {
		drawDate := draw.DrawDate
		detail.DrawDate = &drawDate
		detail.DrawRedNumbers = draw.RedNumbers
		detail.DrawBlueNumbers = draw.BlueNumbers
	}

	return detail
}
