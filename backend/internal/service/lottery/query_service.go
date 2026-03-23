package lottery

import (
	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"
)

type DashboardData struct {
	Lottery              model.LotteryType     `json:"lottery"`
	LatestDraw           *model.DrawResult     `json:"latestDraw"`
	LatestRecommendation *model.Recommendation `json:"latestRecommendation"`
	RecentTickets        []TicketDetail        `json:"recentTickets"`
	Stats                DashboardStats        `json:"stats"`
}

type DashboardStats struct {
	TotalTickets             int     `json:"totalTickets"`
	WonTickets               int     `json:"wonTickets"`
	TotalCost                float64 `json:"totalCost"`
	TotalPrize               float64 `json:"totalPrize"`
	TotalRecommendations     int     `json:"totalRecommendations"`
	PurchasedRecommendations int     `json:"purchasedRecommendations"`
}

func loadDashboardStats(code string, userID string) DashboardStats {
	stats := DashboardStats{}
	var totalTickets int64
	var wonTickets int64

	query := currentUserScope(db.DB.Model(&model.Ticket{}), userID)
	if code != "" {
		query = query.Where("lottery_code = ?", code)
	}
	query.Count(&totalTickets)

	winQuery := currentUserScope(db.DB.Model(&model.Ticket{}), userID).Where("status = ?", TicketStatusWon)
	if code != "" {
		winQuery = winQuery.Where("lottery_code = ?", code)
	}
	winQuery.Count(&wonTickets)

	costQuery := currentUserScope(db.DB.Model(&model.Ticket{}), userID)
	if code != "" {
		costQuery = costQuery.Where("lottery_code = ?", code)
	}
	costQuery.Select("COALESCE(sum(cost_amount), 0)").Scan(&stats.TotalCost)

	prizeQuery := currentUserScope(db.DB.Model(&model.Ticket{}), userID)
	if code != "" {
		prizeQuery = prizeQuery.Where("lottery_code = ?", code)
	}
	prizeQuery.Select("COALESCE(sum(prize_amount), 0)").Scan(&stats.TotalPrize)

	stats.TotalTickets = int(totalTickets)
	stats.WonTickets = int(wonTickets)
	stats.TotalRecommendations = loadRecommendationCount(code, userID)
	stats.PurchasedRecommendations = loadPurchasedRecommendationCount(code, userID)
	return stats
}

func loadRecommendationCount(code string, userID string) int {
	query := currentUserScope(db.DB.Model(&model.Recommendation{}), userID)
	if code != "" {
		query = query.Where("lottery_code = ?", code)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func loadPurchasedRecommendationCount(code string, userID string) int {
	query := currentUserScope(db.DB.Model(&model.Ticket{}), userID).Where("recommendation_id IS NOT NULL")
	if code != "" {
		query = query.Where("lottery_code = ?", code)
	}

	var count int64
	if err := query.Distinct("recommendation_id").Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func getLotteryType(code string) (model.LotteryType, error) {
	lotteryType := model.LotteryType{}
	if err := db.DB.Where("code = ?", code).First(&lotteryType).Error; err != nil {
		return lotteryType, err
	}
	return lotteryType, nil
}

func ListLotteryTypes() ([]model.LotteryType, error) {
	items := make([]model.LotteryType, 0)
	if err := db.DB.Order("created_at asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetLatestRecommendation(code string, userID string) (*model.Recommendation, error) {
	item := model.Recommendation{}
	if err := currentUserScope(db.DB.Preload("Entries"), userID).Where("lottery_code = ?", code).Order("created_at desc").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func GetDashboard(code string, userID string) (*DashboardData, error) {
	lotteryType, err := getLotteryType(code)
	if err != nil {
		return nil, err
	}

	var latestDraw *model.DrawResult
	draw := model.DrawResult{}
	if err := db.DB.Where("lottery_code = ?", code).Order("issue desc").First(&draw).Error; err == nil {
		latestDraw = &draw
	}

	var latestRecommendation *model.Recommendation
	recommendation, err := GetLatestRecommendation(code, userID)
	if err == nil {
		latestRecommendation = recommendation
	}

	recentTickets, err := ListTickets(code, 10, userID)
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		Lottery:              lotteryType,
		LatestDraw:           latestDraw,
		LatestRecommendation: latestRecommendation,
		RecentTickets:        recentTickets,
		Stats:                loadDashboardStats(code, userID),
	}, nil
}

func GetGlobalDashboard(userID string) (*DashboardData, error) {
	return &DashboardData{
		RecentTickets: make([]TicketDetail, 0),
		Stats:         loadDashboardStats("", userID),
	}, nil
}
