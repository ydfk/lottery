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
	TotalTickets int     `json:"totalTickets"`
	WonTickets   int     `json:"wonTickets"`
	TotalCost    float64 `json:"totalCost"`
	TotalPrize   float64 `json:"totalPrize"`
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

func GetLatestRecommendation(code string) (*model.Recommendation, error) {
	item := model.Recommendation{}
	if err := db.DB.Preload("Entries").Where("lottery_code = ?", code).Order("created_at desc").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func GetDashboard(code string) (*DashboardData, error) {
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
	recommendation, err := GetLatestRecommendation(code)
	if err == nil {
		latestRecommendation = recommendation
	}

	recentTickets, err := ListTickets(code, 10)
	if err != nil {
		return nil, err
	}

	stats := DashboardStats{}
	var totalTickets int64
	var wonTickets int64
	db.DB.Model(&model.Ticket{}).Where("lottery_code = ?", code).Count(&totalTickets)
	db.DB.Model(&model.Ticket{}).Where("lottery_code = ? AND status = ?", code, TicketStatusWon).Count(&wonTickets)
	stats.TotalTickets = int(totalTickets)
	stats.WonTickets = int(wonTickets)
	db.DB.Model(&model.Ticket{}).Where("lottery_code = ?", code).Select("COALESCE(sum(cost_amount), 0)").Scan(&stats.TotalCost)
	db.DB.Model(&model.Ticket{}).Where("lottery_code = ?", code).Select("COALESCE(sum(prize_amount), 0)").Scan(&stats.TotalPrize)

	return &DashboardData{
		Lottery:              lotteryType,
		LatestDraw:           latestDraw,
		LatestRecommendation: latestRecommendation,
		RecentTickets:        recentTickets,
		Stats:                stats,
	}, nil
}
