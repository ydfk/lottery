package lottery

import (
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DrawHistoryItem struct {
	Id              uuid.UUID `json:"id"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	LotteryCode     string    `json:"lotteryCode"`
	Issue           string    `json:"issue"`
	DrawDate        time.Time `json:"drawDate"`
	RedNumbers      string    `json:"redNumbers"`
	BlueNumbers     string    `json:"blueNumbers"`
	SaleAmount      float64   `json:"saleAmount"`
	PrizePoolAmount float64   `json:"prizePoolAmount"`
	Source          string    `json:"source"`
}

type DrawQueryOptions struct {
	Page        int
	PageSize    int
	LotteryCode string
	Sort        string
}

type DrawPageResult struct {
	Items    []DrawHistoryItem `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	Total    int64             `json:"total"`
	HasMore  bool              `json:"hasMore"`
}

func QueryDrawResults(options DrawQueryOptions) (*DrawPageResult, error) {
	page := max(1, options.Page)
	pageSize := options.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query := applyDrawFilters(db.DB.Model(&model.DrawResult{}), options)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	items := make([]DrawHistoryItem, 0)
	if err := applyDrawSort(query, options.Sort).
		Select(
			"id",
			"created_at",
			"updated_at",
			"lottery_code",
			"issue",
			"draw_date",
			"red_numbers",
			"blue_numbers",
			"sale_amount",
			"prize_pool_amount",
			"source",
		).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &DrawPageResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasMore:  int64(page*pageSize) < total,
	}, nil
}

func applyDrawFilters(query *gorm.DB, options DrawQueryOptions) *gorm.DB {
	if options.LotteryCode != "" {
		query = query.Where("lottery_code = ?", options.LotteryCode)
	}
	return query
}

func applyDrawSort(query *gorm.DB, sort string) *gorm.DB {
	switch sort {
	case "oldest":
		return query.Order("draw_date asc").Order("issue asc").Order("created_at asc")
	default:
		return query.Order("draw_date desc").Order("issue desc").Order("created_at desc")
	}
}
