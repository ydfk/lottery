package lottery

import (
	"errors"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

type RecommendationDetail struct {
	model.Recommendation
	EntryCount      int           `json:"entryCount"`
	WinningCount    int           `json:"winningCount"`
	IsPurchased     bool          `json:"isPurchased"`
	PurchasedTicket *TicketDetail `json:"purchasedTicket,omitempty"`
}

func ListRecommendations(code string, limit int) ([]RecommendationDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	recommendations := make([]model.Recommendation, 0)
	if err := db.DB.Preload("Entries").Where("lottery_code = ?", code).Order("created_at desc").Limit(limit).Find(&recommendations).Error; err != nil {
		return nil, err
	}

	result := make([]RecommendationDetail, 0, len(recommendations))
	for _, recommendation := range recommendations {
		detail, err := buildRecommendationDetail(recommendation)
		if err != nil {
			return nil, err
		}
		result = append(result, *detail)
	}
	return result, nil
}

func GetRecommendationDetail(code string, recommendationID string) (*RecommendationDetail, error) {
	recommendation := model.Recommendation{}
	if err := db.DB.Preload("Entries").First(&recommendation, "id = ? AND lottery_code = ?", recommendationID, code).Error; err != nil {
		return nil, err
	}
	return buildRecommendationDetail(recommendation)
}

func buildRecommendationDetail(recommendation model.Recommendation) (*RecommendationDetail, error) {
	detail := &RecommendationDetail{
		Recommendation: recommendation,
		EntryCount:     len(recommendation.Entries),
	}

	for _, entry := range recommendation.Entries {
		if entry.IsWinning {
			detail.WinningCount++
		}
	}

	ticket := model.Ticket{}
	err := db.DB.Preload("Entries").
		Where("lottery_code = ? AND recommendation_id = ?", recommendation.LotteryCode, recommendation.Id).
		Order("created_at desc").
		First(&ticket).Error
	if err == nil {
		detail.IsPurchased = true
		detail.PurchasedTicket = &TicketDetail{
			Ticket:   ticket,
			ImageURL: buildPublicImageURL(ticket.ImagePath),
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return detail, nil
}
