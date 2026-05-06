package lottery

import (
	"fmt"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DrawHistoryItem struct {
	Id                uuid.UUID       `json:"id"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	LotteryCode       string          `json:"lotteryCode"`
	Issue             string          `json:"issue"`
	DrawDate          time.Time       `json:"drawDate"`
	RedNumbers        string          `json:"redNumbers"`
	BlueNumbers       string          `json:"blueNumbers"`
	SaleAmount        float64         `json:"saleAmount"`
	PrizePoolAmount   float64         `json:"prizePoolAmount"`
	FirstPrizeAmount  float64         `json:"firstPrizeAmount"`
	SecondPrizeAmount float64         `json:"secondPrizeAmount"`
	Source            string          `json:"source"`
	RawPayload        string          `json:"rawPayload"`
	PrizeDetails      []DrawPrizeItem `json:"prizeDetails"`
}

type DrawPrizeItem struct {
	Id          uuid.UUID `json:"id"`
	PrizeName   string    `json:"prizeName"`
	PrizeRule   string    `json:"prizeRule"`
	WinnerCount int       `json:"winnerCount"`
	SingleBonus float64   `json:"singleBonus"`
}

type DrawQueryOptions struct {
	Page        int
	PageSize    int
	LotteryCode string
	Issue       string
	DrawDate    string
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
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query := applyDrawFilters(db.DB.Model(&model.DrawResult{}), options)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	draws := make([]model.DrawResult, 0)
	if err := applyDrawSort(query, options.Sort).
		Preload("PrizeDetails", func(query *gorm.DB) *gorm.DB {
			return query.Order("created_at asc")
		}).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&draws).Error; err != nil {
		return nil, err
	}

	items := make([]DrawHistoryItem, 0, len(draws))
	for _, draw := range draws {
		items = append(items, buildDrawHistoryItem(draw))
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
	if strings.TrimSpace(options.Issue) != "" {
		query = query.Where("issue IN ?", buildDrawIssueAliases(options.LotteryCode, options.Issue))
	}
	if strings.TrimSpace(options.DrawDate) != "" {
		drawDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(options.DrawDate), time.Local)
		if err != nil {
			query.AddError(fmt.Errorf("开奖日期格式不正确"))
			return query
		}
		query = query.Where("draw_date >= ? AND draw_date < ?", drawDate, drawDate.AddDate(0, 0, 1))
	}
	return query
}

func buildDrawIssueAliases(code string, issue string) []string {
	aliases := issueAliases(code, issue)
	if code == "" {
		for _, item := range issueAliases("dlt", issue) {
			exists := false
			for _, current := range aliases {
				if current == item {
					exists = true
					break
				}
			}
			if !exists {
				aliases = append(aliases, item)
			}
		}
	}
	return aliases
}

func applyDrawSort(query *gorm.DB, sort string) *gorm.DB {
	switch sort {
	case "oldest":
		return query.Order("draw_date asc").Order("issue asc").Order("created_at asc")
	default:
		return query.Order("draw_date desc").Order("issue desc").Order("created_at desc")
	}
}

func buildDrawHistoryItem(draw model.DrawResult) DrawHistoryItem {
	item := DrawHistoryItem{
		Id:              draw.Id,
		CreatedAt:       draw.CreatedAt,
		UpdatedAt:       draw.UpdatedAt,
		LotteryCode:     draw.LotteryCode,
		Issue:           draw.Issue,
		DrawDate:        draw.DrawDate,
		RedNumbers:      draw.RedNumbers,
		BlueNumbers:     draw.BlueNumbers,
		SaleAmount:      draw.SaleAmount,
		PrizePoolAmount: draw.PrizePoolAmount,
		Source:          draw.Source,
		RawPayload:      draw.RawPayload,
		PrizeDetails:    make([]DrawPrizeItem, 0, len(draw.PrizeDetails)),
	}

	for _, prize := range draw.PrizeDetails {
		prizeItem := DrawPrizeItem{
			Id:          prize.Id,
			PrizeName:   prize.PrizeName,
			PrizeRule:   prize.PrizeRule,
			WinnerCount: prize.WinnerCount,
			SingleBonus: prize.SingleBonus,
		}
		item.PrizeDetails = append(item.PrizeDetails, prizeItem)

		if isPrizeName(prize.PrizeName, "一等奖") && prize.SingleBonus > item.FirstPrizeAmount {
			item.FirstPrizeAmount = prize.SingleBonus
		}
		if isPrizeName(prize.PrizeName, "二等奖") && prize.SingleBonus > item.SecondPrizeAmount {
			item.SecondPrizeAmount = prize.SingleBonus
		}
	}

	return item
}

func isPrizeName(value string, target string) bool {
	normalized := normalizePrizeName(value)
	return normalized == target || strings.Contains(value, target)
}
