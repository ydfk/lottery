package lottery

import (
	"errors"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"
)

const (
	DashboardTrendRange30Days = "30d"
	DashboardTrendRange90Days = "90d"
	DashboardTrendRange1Year  = "1y"
)

var chinaStandardTime = time.FixedZone("Asia/Shanghai", 8*60*60)

var ErrInvalidDashboardTrendRange = errors.New("趋势范围不正确，可选 30d、90d、1y")

type DashboardTrend struct {
	Range  string                `json:"range"`
	Bucket string                `json:"bucket"`
	Items  []DashboardTrendPoint `json:"items"`
}

type DashboardTrendPoint struct {
	PeriodStart         string  `json:"periodStart"`
	TotalCost           float64 `json:"totalCost"`
	TotalPrize          float64 `json:"totalPrize"`
	TicketCount         int     `json:"ticketCount"`
	WonTickets          int     `json:"wonTickets"`
	RecommendationCount int     `json:"recommendationCount"`
}

type dashboardTrendWindow struct {
	rangeValue string
	bucket     string
	starts     []time.Time
	end        time.Time
}

func GetDashboardTrend(userID string, rangeValue string) (*DashboardTrend, error) {
	return buildDashboardTrend(userID, rangeValue, time.Now())
}

func buildDashboardTrend(userID string, rangeValue string, now time.Time) (*DashboardTrend, error) {
	window, err := makeDashboardTrendWindow(rangeValue, now.In(chinaStandardTime))
	if err != nil {
		return nil, err
	}

	items := make([]DashboardTrendPoint, len(window.starts))
	indexes := make(map[string]int, len(window.starts))
	for index, start := range window.starts {
		key := start.Format("2006-01-02")
		indexes[key] = index
		items[index].PeriodStart = key
	}

	tickets := make([]model.Ticket, 0)
	if err := currentUserScope(db.DB.Model(&model.Ticket{}), userID).
		Select("purchased_at", "cost_amount", "prize_amount", "status").
		Where("purchased_at >= ? AND purchased_at <= ?", window.starts[0], window.end).
		Find(&tickets).Error; err != nil {
		return nil, err
	}
	for _, ticket := range tickets {
		key := dashboardTrendBucketStart(ticket.PurchasedAt.In(chinaStandardTime), window.bucket).Format("2006-01-02")
		index, exists := indexes[key]
		if !exists {
			continue
		}
		items[index].TotalCost += ticket.CostAmount
		items[index].TotalPrize += ticket.PrizeAmount
		items[index].TicketCount++
		if ticket.Status == TicketStatusWon {
			items[index].WonTickets++
		}
	}

	recommendations := make([]model.Recommendation, 0)
	if err := currentUserScope(db.DB.Model(&model.Recommendation{}), userID).
		Select("created_at").
		Where("created_at >= ? AND created_at <= ?", window.starts[0], window.end).
		Find(&recommendations).Error; err != nil {
		return nil, err
	}
	for _, recommendation := range recommendations {
		key := dashboardTrendBucketStart(recommendation.CreatedAt.In(chinaStandardTime), window.bucket).Format("2006-01-02")
		if index, exists := indexes[key]; exists {
			items[index].RecommendationCount++
		}
	}

	return &DashboardTrend{Range: window.rangeValue, Bucket: window.bucket, Items: items}, nil
}

func makeDashboardTrendWindow(rangeValue string, now time.Time) (dashboardTrendWindow, error) {
	var bucket string
	var starts []time.Time
	switch rangeValue {
	case DashboardTrendRange30Days:
		bucket = "day"
		first := startOfChinaDay(now).AddDate(0, 0, -29)
		starts = make([]time.Time, 30)
		for index := range starts {
			starts[index] = first.AddDate(0, 0, index)
		}
	case DashboardTrendRange90Days:
		bucket = "week"
		first := startOfChinaWeek(now).AddDate(0, 0, -12*7)
		starts = make([]time.Time, 13)
		for index := range starts {
			starts[index] = first.AddDate(0, 0, index*7)
		}
	case DashboardTrendRange1Year:
		bucket = "month"
		currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, chinaStandardTime)
		first := currentMonth.AddDate(0, -11, 0)
		starts = make([]time.Time, 12)
		for index := range starts {
			starts[index] = first.AddDate(0, index, 0)
		}
	default:
		return dashboardTrendWindow{}, ErrInvalidDashboardTrendRange
	}

	return dashboardTrendWindow{
		rangeValue: rangeValue,
		bucket:     bucket,
		starts:     starts,
		end:        now,
	}, nil
}

func dashboardTrendBucketStart(value time.Time, bucket string) time.Time {
	switch bucket {
	case "week":
		return startOfChinaWeek(value)
	case "month":
		return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, chinaStandardTime)
	default:
		return startOfChinaDay(value)
	}
}

func startOfChinaDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, chinaStandardTime)
}

func startOfChinaWeek(value time.Time) time.Time {
	day := startOfChinaDay(value)
	offset := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -offset)
}
