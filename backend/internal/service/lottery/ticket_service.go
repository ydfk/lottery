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
	ImageURL string `json:"imageUrl"`
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

func recognizeTicket(ctx context.Context, definition Definition, lotteryType model.LotteryType, imagePath string, ocrText string) (*RecognitionResult, error) {
	if ocrText != "" {
		if definition.Code != "ssq" {
			return nil, fmt.Errorf("当前仅支持双色球 OCR 文本识别")
		}
		return ParseSSQText(ocrText)
	}

	recognizer := newVisionRecognizer(resolveValue(config.Current.Vision.Provider, lotteryType.VisionProvider))
	if recognizer == nil {
		return nil, fmt.Errorf("未配置视觉模型，请填写 OCR 文本作为降级输入")
	}

	recognized, err := recognizer.Recognize(ctx, lotteryType, imagePath)
	if err != nil {
		return nil, err
	}
	if len(recognized.Entries) == 0 && recognized.RawText != "" {
		if definition.Code != "ssq" {
			return nil, fmt.Errorf("当前仅支持双色球自动识别")
		}
		return ParseSSQText(recognized.RawText)
	}
	if len(recognized.Entries) == 0 {
		return nil, fmt.Errorf("未从图片中识别到有效号码")
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
	if err := db.DB.Where("lottery_code = ? AND issue = ?", code, issue).Find(&tickets).Error; err != nil {
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

	draw := model.DrawResult{}
	if err := db.DB.Preload("PrizeDetails").Where("lottery_code = ? AND issue = ?", ticket.LotteryCode, ticket.Issue).First(&draw).Error; err != nil {
		return nil
	}

	prizeMap := make(map[string]float64, len(draw.PrizeDetails))
	for _, prize := range draw.PrizeDetails {
		prizeMap[normalizePrizeName(prize.PrizeName)] = prize.SingleBonus
	}

	totalPrize := 0.0
	hasWinning := false
	for _, entry := range ticket.Entries {
		result := JudgeNumbers(ticket.LotteryCode, entry.RedNumbers, entry.BlueNumbers, draw, prizeMap)
		entry.IsWinning = result.IsWinning
		entry.PrizeName = result.PrizeName
		entry.PrizeAmount = result.PrizeAmount
		entry.MatchSummary = result.MatchSummary
		totalPrize += result.PrizeAmount
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
	return db.DB.Save(&ticket).Error
}

func GetTicketDetail(ticketID string) (*TicketDetail, error) {
	ticket := model.Ticket{}
	if err := db.DB.Preload("Entries").First(&ticket, "id = ?", ticketID).Error; err != nil {
		return nil, err
	}
	return &TicketDetail{
		Ticket:   ticket,
		ImageURL: buildPublicImageURL(ticket.ImagePath),
	}, nil
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
		result = append(result, TicketDetail{
			Ticket:   ticket,
			ImageURL: buildPublicImageURL(ticket.ImagePath),
		})
	}
	return result, nil
}
