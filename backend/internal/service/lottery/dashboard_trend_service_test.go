package lottery

import (
	"testing"
	"time"

	"go-fiber-starter/internal/model/base"
	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
)

func TestDashboardTrendAggregatesAndScopesByUser(t *testing.T) {
	setupImportTicketTestDB(t)
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, chinaStandardTime)
	userID := uuid.New()
	otherUserID := uuid.New()

	createTrendTicket(t, userID, now.AddDate(0, 0, -2), 10, 25, TicketStatusWon)
	createTrendTicket(t, userID, time.Date(2026, 7, 24, 0, 30, 0, 0, chinaStandardTime), 4, 0, TicketStatusPending)
	createTrendTicket(t, userID, now.AddDate(0, 0, -35), 20, 0, TicketStatusPending)
	createTrendTicket(t, otherUserID, now.AddDate(0, 0, -2), 999, 999, TicketStatusWon)
	createTrendRecommendation(t, userID, now.AddDate(0, 0, -2))
	createTrendRecommendation(t, otherUserID, now.AddDate(0, 0, -2))

	trend, err := buildDashboardTrend(userID.String(), DashboardTrendRange30Days, now)
	if err != nil {
		t.Fatalf("build dashboard trend: %v", err)
	}
	if len(trend.Items) != 30 || trend.Bucket != "day" {
		t.Fatalf("unexpected trend window: bucket=%s items=%d", trend.Bucket, len(trend.Items))
	}

	today := trendPointByDate(t, trend.Items, "2026-07-24")
	if today.TotalCost != 4 || today.TicketCount != 1 {
		t.Fatalf("china date boundary mismatch: %+v", today)
	}
	point := trendPointByDate(t, trend.Items, "2026-07-22")
	if point.TotalCost != 10 || point.TotalPrize != 25 || point.TicketCount != 1 || point.WonTickets != 1 {
		t.Fatalf("ticket aggregation mismatch: %+v", point)
	}
	if point.RecommendationCount != 1 {
		t.Fatalf("recommendation aggregation mismatch: %+v", point)
	}
	if trendPointByDate(t, trend.Items, "2026-07-23").TicketCount != 0 {
		t.Fatal("empty bucket should be filled with zero")
	}
}

func TestDashboardTrendRangesAndValidation(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, chinaStandardTime)

	weekly, err := makeDashboardTrendWindow(DashboardTrendRange90Days, now)
	if err != nil || weekly.bucket != "week" || len(weekly.starts) != 13 {
		t.Fatalf("unexpected 90-day window: %+v err=%v", weekly, err)
	}
	monthly, err := makeDashboardTrendWindow(DashboardTrendRange1Year, now)
	if err != nil || monthly.bucket != "month" || len(monthly.starts) != 12 {
		t.Fatalf("unexpected one-year window: %+v err=%v", monthly, err)
	}
	if _, err := makeDashboardTrendWindow("all", now); err == nil {
		t.Fatal("invalid range should return an error")
	}
}

func createTrendTicket(t *testing.T, userID uuid.UUID, purchasedAt time.Time, cost float64, prize float64, status string) {
	t.Helper()
	ticket := model.Ticket{
		UserID:      &userID,
		LotteryCode: "ssq",
		Issue:       uuid.NewString(),
		Status:      status,
		CostAmount:  cost,
		PrizeAmount: prize,
		PurchasedAt: purchasedAt,
	}
	if err := db.DB.Create(&ticket).Error; err != nil {
		t.Fatalf("create trend ticket: %v", err)
	}
}

func createTrendRecommendation(t *testing.T, userID uuid.UUID, createdAt time.Time) {
	t.Helper()
	recommendation := model.Recommendation{
		BaseModel:   base.BaseModel{CreatedAt: createdAt, UpdatedAt: createdAt},
		UserID:      &userID,
		LotteryCode: "ssq",
		Issue:       uuid.NewString(),
	}
	if err := db.DB.Create(&recommendation).Error; err != nil {
		t.Fatalf("create trend recommendation: %v", err)
	}
}

func trendPointByDate(t *testing.T, items []DashboardTrendPoint, date string) DashboardTrendPoint {
	t.Helper()
	for _, item := range items {
		if item.PeriodStart == date {
			return item
		}
	}
	t.Fatalf("trend point not found: %s", date)
	return DashboardTrendPoint{}
}
