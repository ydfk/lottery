package lottery

import (
	"errors"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

type RecommendationDetail struct {
	model.Recommendation
	EntryCount       int            `json:"entryCount"`
	WinningCount     int            `json:"winningCount"`
	IsPurchased      bool           `json:"isPurchased"`
	PurchasedCount   int            `json:"purchasedCount"`
	PurchasedTickets []TicketDetail `json:"purchasedTickets"`
}

type RecommendationQueryOptions struct {
	UserID      string
	Page        int
	PageSize    int
	LotteryCode string
	Status      string
	Sort        string
}

type RecommendationPageResult struct {
	Items    []RecommendationDetail `json:"items"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
	Total    int64                  `json:"total"`
	HasMore  bool                   `json:"hasMore"`
}

func ListRecommendations(code string, limit int, userID string) ([]RecommendationDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	recommendations := make([]model.Recommendation, 0)
	if err := currentUserScope(db.DB.Preload("Entries"), userID).Where("lottery_code = ?", code).Order("created_at desc").Limit(limit).Find(&recommendations).Error; err != nil {
		return nil, err
	}

	result := make([]RecommendationDetail, 0, len(recommendations))
	for _, recommendation := range recommendations {
		detail, err := buildRecommendationDetail(recommendation, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, *detail)
	}
	return result, nil
}

func ListAllRecommendations(limit int, userID string) ([]RecommendationDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	recommendations := make([]model.Recommendation, 0)
	if err := currentUserScope(db.DB.Preload("Entries"), userID).Order("created_at desc").Limit(limit).Find(&recommendations).Error; err != nil {
		return nil, err
	}

	result := make([]RecommendationDetail, 0, len(recommendations))
	for _, recommendation := range recommendations {
		detail, err := buildRecommendationDetail(recommendation, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, *detail)
	}
	return result, nil
}

func QueryRecommendations(options RecommendationQueryOptions) (*RecommendationPageResult, error) {
	page := max(1, options.Page)
	pageSize := options.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query := applyRecommendationFilters(currentUserScope(db.DB.Model(&model.Recommendation{}), options.UserID), options)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	recommendations := make([]model.Recommendation, 0)
	if err := applyRecommendationSort(query, options.Sort).
		Preload("Entries").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&recommendations).Error; err != nil {
		return nil, err
	}

	items := make([]RecommendationDetail, 0, len(recommendations))
	for _, recommendation := range recommendations {
		detail, err := buildRecommendationDetail(recommendation, options.UserID)
		if err != nil {
			return nil, err
		}
		items = append(items, *detail)
	}

	return &RecommendationPageResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  int64(page*pageSize) < total,
	}, nil
}

func GetRecommendationDetail(code string, recommendationID string, userID string) (*RecommendationDetail, error) {
	recommendation := model.Recommendation{}
	if err := currentUserScope(db.DB.Preload("Entries"), userID).First(&recommendation, "id = ? AND lottery_code = ?", recommendationID, code).Error; err != nil {
		return nil, err
	}
	return buildRecommendationDetail(recommendation, userID)
}

func buildRecommendationDetail(recommendation model.Recommendation, userID string) (*RecommendationDetail, error) {
	detail := &RecommendationDetail{
		Recommendation: recommendation,
		EntryCount:     len(recommendation.Entries),
	}

	for _, entry := range recommendation.Entries {
		if entry.IsWinning {
			detail.WinningCount++
		}
	}

	tickets := make([]model.Ticket, 0)
	err := currentUserScope(db.DB.Preload("Entries"), userID).
		Where("lottery_code = ? AND recommendation_id = ?", recommendation.LotteryCode, recommendation.Id).
		Order("created_at desc").
		Find(&tickets).Error
	if err == nil && len(tickets) > 0 {
		detail.IsPurchased = true
		detail.PurchasedCount = len(tickets)
		detail.PurchasedTickets = make([]TicketDetail, 0, len(tickets))
		for _, ticket := range tickets {
			detail.PurchasedTickets = append(detail.PurchasedTickets, *buildTicketDetail(ticket))
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return detail, nil
}

func applyRecommendationFilters(query *gorm.DB, options RecommendationQueryOptions) *gorm.DB {
	if options.LotteryCode != "" {
		query = query.Where("lottery_code = ?", options.LotteryCode)
	}

	switch options.Status {
	case TicketStatusPending:
		query = query.Where("checked_at IS NULL")
	case TicketStatusWon:
		query = query.Where("checked_at IS NOT NULL").Where("prize_amount > 0")
	case TicketStatusNotWon:
		query = query.Where("checked_at IS NOT NULL").Where("prize_amount <= 0")
	}

	return query
}

func applyRecommendationSort(query *gorm.DB, sort string) *gorm.DB {
	switch sort {
	case "oldest":
		return query.Order("created_at asc").Order("id asc")
	case "draw_latest":
		return query.Order("draw_date desc").Order("created_at desc").Order("id desc")
	case "draw_oldest":
		return query.Order("draw_date asc").Order("created_at asc").Order("id asc")
	case "prize_high":
		return query.Order("prize_amount desc").Order("created_at desc").Order("id desc")
	default:
		return query.Order("created_at desc").Order("id desc")
	}
}
