package lottery

import (
	"testing"
	"time"

	model "go-fiber-starter/internal/model/lottery"
)

func TestBuildRecommendationPlanForDLT(t *testing.T) {
	now := time.Date(2026, 3, 10, 10, 0, 0, 0, time.Local)
	definition := Definition{
		Code: "dlt",
		Name: "体彩大乐透",
		DrawSchedule: DrawScheduleSettings{
			Weekdays: []int{1, 2, 6},
			Time:     "21:30",
		},
	}
	history := []model.DrawResult{
		{
			LotteryCode: "dlt",
			Issue:       "2026025",
			DrawDate:    time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local),
		},
	}

	issue, drawDate, err := buildRecommendationPlan(definition, history, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue != "2026026" {
		t.Fatalf("unexpected issue: %s", issue)
	}
	if drawDate.Format("2006-01-02 15:04") != "2026-03-10 21:30" {
		t.Fatalf("unexpected draw date: %s", drawDate.Format("2006-01-02 15:04"))
	}
}

func TestBuildRecommendationPlanForSSQ(t *testing.T) {
	now := time.Date(2026, 3, 17, 20, 0, 0, 0, time.Local)
	definition := Definition{
		Code: "ssq",
		Name: "福彩双色球",
		DrawSchedule: DrawScheduleSettings{
			Weekdays: []int{0, 2, 4},
			Time:     "21:30",
		},
	}
	history := []model.DrawResult{
		{
			LotteryCode: "ssq",
			Issue:       "2026028",
			DrawDate:    time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
		},
	}

	issue, drawDate, err := buildRecommendationPlan(definition, history, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue != "2026029" {
		t.Fatalf("unexpected issue: %s", issue)
	}
	if drawDate.Format("2006-01-02 15:04") != "2026-03-17 21:30" {
		t.Fatalf("unexpected draw date: %s", drawDate.Format("2006-01-02 15:04"))
	}
}
