package lottery

import (
	"context"
	"errors"
	"fmt"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

type ScanTicketInput struct {
	UserID      string
	Code        string
	Issue       string
	ImagePath   string
	OCRText     string
	PurchasedAt time.Time
	Notes       string
}

type TicketDetail struct {
	model.Ticket
	ImageURL        string                      `json:"imageUrl"`
	DrawDate        *time.Time                  `json:"drawDate"`
	DrawRedNumbers  string                      `json:"drawRedNumbers"`
	DrawBlueNumbers string                      `json:"drawBlueNumbers"`
	Recommendation  *TicketRecommendationDetail `json:"recommendation,omitempty"`
}

type TicketRecommendationDetail struct {
	ID        string                      `json:"id"`
	Issue     string                      `json:"issue"`
	DrawDate  *time.Time                  `json:"drawDate,omitempty"`
	Summary   string                      `json:"summary"`
	Entries   []TicketRecommendationEntry `json:"entries"`
	CreatedAt time.Time                   `json:"createdAt"`
}

type TicketRecommendationEntry struct {
	ID          string  `json:"id"`
	Sequence    int     `json:"sequence"`
	RedNumbers  string  `json:"redNumbers"`
	BlueNumbers string  `json:"blueNumbers"`
	PrizeAmount float64 `json:"prizeAmount"`
	PrizeName   string  `json:"prizeName"`
}

type TicketQueryOptions struct {
	UserID      string
	Page        int
	PageSize    int
	LotteryCode string
	Status      string
	Sort        string
}

type TicketPageResult struct {
	Items    []TicketDetail `json:"items"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Total    int64          `json:"total"`
	HasMore  bool           `json:"hasMore"`
}

func ScanTicket(ctx context.Context, input ScanTicketInput) (*TicketDetail, error) {
	upload, err := UploadTicketImage(UploadTicketImageInput{
		UserID:           input.UserID,
		Code:             input.Code,
		ImagePath:        input.ImagePath,
		OriginalFilename: "",
	})
	if err != nil {
		return nil, err
	}

	recognized, err := RecognizeUploadedTicket(ctx, RecognizeUploadedTicketInput{
		UserID:   input.UserID,
		Code:     input.Code,
		UploadID: upload.Id.String(),
		OCRText:  input.OCRText,
	})
	if err != nil {
		return nil, err
	}

	return CreateTicket(ctx, CreateTicketInput{
		UserID:      input.UserID,
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
	if len(ticket.Entries) == 0 {
		return nil
	}

	canonicalIssue := normalizeIssueByCode(ticket.LotteryCode, ticket.Issue)
	if canonicalIssue != "" && canonicalIssue != ticket.Issue {
		ticket.Issue = canonicalIssue
		if err := db.DB.Omit("Entries").Save(&ticket).Error; err != nil {
			return err
		}
	}

	draw, err := findSettlementDraw(ticket.LotteryCode, ticket.Issue)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if shouldDeferSettlement(ticket.LotteryCode, ticket.ManualDrawDate) {
			return resetTicketPending(ticket.Id.String())
		}
		return nil
	}
	if err != nil {
		return err
	}

	prizeMap := make(map[string]float64, len(draw.PrizeDetails))
	for _, prize := range draw.PrizeDetails {
		prizeMap[normalizePrizeName(prize.PrizeName)] = prize.SingleBonus
	}

	totalPrize := 0.0
	hasWinning := false
	for _, entry := range ticket.Entries {
		result := JudgeNumbers(ticket.LotteryCode, entry.RedNumbers, entry.BlueNumbers, entry.IsAdditional, *draw, prizeMap)
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

func RecheckTicket(ctx context.Context, ticketID string, code string, userID string) (*TicketDetail, error) {
	ticket := model.Ticket{}
	query := currentUserScope(db.DB.Preload("Entries"), userID)
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
	if _, err := findSettlementDraw(ticket.LotteryCode, ticketIssue); errors.Is(err, gorm.ErrRecordNotFound) {
		if shouldDeferSettlement(ticket.LotteryCode, ticket.ManualDrawDate) {
			if resetErr := resetTicketPending(ticket.Id.String()); resetErr != nil {
				return nil, resetErr
			}
			return GetTicketDetail(ticket.Id.String(), userID)
		}
		if _, syncErr := SyncLatestDraw(ctx, ticket.LotteryCode, ticketIssue); syncErr != nil {
			return nil, syncErr
		}
	} else if err != nil {
		return nil, err
	}
	if err := EvaluateTicket(ticket.Id.String()); err != nil {
		return nil, err
	}
	return GetTicketDetail(ticket.Id.String(), userID)
}

func GetTicketDetail(ticketID string, userID string) (*TicketDetail, error) {
	ticket := model.Ticket{}
	if err := currentUserScope(db.DB.Preload("Entries"), userID).First(&ticket, "id = ?", ticketID).Error; err != nil {
		return nil, err
	}
	return buildTicketDetail(ticket), nil
}

func ListTickets(code string, limit int, userID string) ([]TicketDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	tickets := make([]model.Ticket, 0)
	if err := currentUserScope(db.DB.Preload("Entries"), userID).Where("lottery_code = ?", code).Order("created_at desc").Limit(limit).Find(&tickets).Error; err != nil {
		return nil, err
	}

	result := make([]TicketDetail, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, *buildTicketDetail(ticket))
	}
	return result, nil
}

func ListAllTickets(limit int, userID string) ([]TicketDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	tickets := make([]model.Ticket, 0)
	if err := currentUserScope(db.DB.Preload("Entries"), userID).Order("created_at desc").Limit(limit).Find(&tickets).Error; err != nil {
		return nil, err
	}

	result := make([]TicketDetail, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, *buildTicketDetail(ticket))
	}
	return result, nil
}

func QueryAllTickets(options TicketQueryOptions) (*TicketPageResult, error) {
	page := max(1, options.Page)
	pageSize := options.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query := currentUserScope(db.DB.Model(&model.Ticket{}), options.UserID)
	if options.LotteryCode != "" {
		query = query.Where("lottery_code = ?", options.LotteryCode)
	}
	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	tickets := make([]model.Ticket, 0)
	if err := currentUserScope(db.DB.Preload("Entries"), options.UserID).
		Scopes(applyTicketFilters(options), applyTicketSort(options.Sort)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tickets).Error; err != nil {
		return nil, err
	}

	items := make([]TicketDetail, 0, len(tickets))
	for _, ticket := range tickets {
		items = append(items, *buildTicketDetail(ticket))
	}

	return &TicketPageResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  int64(page*pageSize) < total,
	}, nil
}

func applyTicketFilters(options TicketQueryOptions) func(*gorm.DB) *gorm.DB {
	return func(query *gorm.DB) *gorm.DB {
		if options.LotteryCode != "" {
			query = query.Where("lottery_code = ?", options.LotteryCode)
		}
		if options.Status != "" {
			query = query.Where("status = ?", options.Status)
		}
		return query
	}
}

func applyTicketSort(sort string) func(*gorm.DB) *gorm.DB {
	return func(query *gorm.DB) *gorm.DB {
		switch sort {
		case "oldest":
			return query.Order("purchased_at asc").Order("created_at asc")
		case "prize_high":
			return query.Order("prize_amount desc").Order("purchased_at desc")
		case "cost_high":
			return query.Order("cost_amount desc").Order("purchased_at desc")
		default:
			return query.Order("purchased_at desc").Order("created_at desc")
		}
	}
}

func buildTicketDetail(ticket model.Ticket) *TicketDetail {
	detail := &TicketDetail{
		Ticket:   ticket,
		ImageURL: buildPublicImageURL(ticket.ImagePath),
	}

	if draw, err := findTicketDisplayDraw(ticket.LotteryCode, ticket.Issue); err == nil {
		drawDate := draw.DrawDate
		if ticket.ManualDrawDate != nil && !ticket.ManualDrawDate.IsZero() {
			drawDate = *ticket.ManualDrawDate
		}
		detail.DrawDate = &drawDate
		detail.DrawRedNumbers = draw.RedNumbers
		detail.DrawBlueNumbers = draw.BlueNumbers
	} else if ticket.ManualDrawDate != nil {
		detail.DrawDate = ticket.ManualDrawDate
	}
	if ticket.RecommendationID != nil {
		recommendation := model.Recommendation{}
		query := db.DB.Preload("Entries").Where("id = ?", *ticket.RecommendationID)
		if ticket.UserID != nil {
			query = query.Where("user_id = ?", *ticket.UserID)
		}
		if err := query.First(&recommendation).Error; err == nil {
			detail.Recommendation = buildTicketRecommendationDetail(recommendation)
		}
	}

	return detail
}

func findSettlementDraw(code string, issue string) (*model.DrawResult, error) {
	draw := model.DrawResult{}
	if err := db.DB.Preload("PrizeDetails").
		Where("lottery_code = ? AND issue IN ?", code, issueAliases(code, issue)).
		Order("created_at desc").
		Order("issue desc").
		First(&draw).Error; err != nil {
		return nil, err
	}
	return &draw, nil
}

func findTicketDisplayDraw(code string, issue string) (*model.DrawResult, error) {
	draw := model.DrawResult{}
	if err := db.DB.Select("issue", "draw_date", "red_numbers", "blue_numbers").
		Where("lottery_code = ? AND issue IN ?", code, issueAliases(code, issue)).
		Order("created_at desc").
		Order("issue desc").
		First(&draw).Error; err != nil {
		return nil, err
	}
	return &draw, nil
}

func resetTicketPending(ticketID string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.TicketEntry{}).
			Where("ticket_id = ?", ticketID).
			Updates(map[string]any{
				"is_winning":    false,
				"prize_name":    "",
				"prize_amount":  0,
				"match_summary": "待开奖",
			}).Error; err != nil {
			return err
		}

		return tx.Model(&model.Ticket{}).
			Where("id = ?", ticketID).
			Updates(map[string]any{
				"status":       TicketStatusPending,
				"checked_at":   nil,
				"prize_amount": 0,
			}).Error
	})
}

func buildTicketRecommendationDetail(recommendation model.Recommendation) *TicketRecommendationDetail {
	entries := make([]TicketRecommendationEntry, 0, len(recommendation.Entries))
	for _, entry := range recommendation.Entries {
		entries = append(entries, TicketRecommendationEntry{
			ID:          entry.Id.String(),
			Sequence:    entry.Sequence,
			RedNumbers:  entry.RedNumbers,
			BlueNumbers: entry.BlueNumbers,
			PrizeAmount: entry.PrizeAmount,
			PrizeName:   entry.PrizeName,
		})
	}

	return &TicketRecommendationDetail{
		ID:        recommendation.Id.String(),
		Issue:     recommendation.Issue,
		DrawDate:  recommendation.DrawDate,
		Summary:   recommendation.Summary,
		Entries:   entries,
		CreatedAt: recommendation.CreatedAt,
	}
}
