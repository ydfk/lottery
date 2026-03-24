package lottery

import (
	"testing"
	"time"
)

func TestBuildRecommendationPlanForDLT(t *testing.T) {
	now := time.Date(2026, 3, 21, 23, 0, 0, 0, time.Local)
	definition := Definition{
		Code: "dlt",
		Name: "体彩大乐透",
		DrawSchedule: DrawScheduleSettings{
			Weekdays: []int{1, 2, 6},
			Time:     "21:30",
		},
	}
	issue, drawDate, err := buildRecommendationPlanFromAnchor(
		definition,
		"2026029",
		time.Date(2026, 3, 21, 0, 0, 0, 0, time.Local),
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue != "2026030" {
		t.Fatalf("unexpected issue: %s", issue)
	}
	if drawDate.Format("2006-01-02 15:04") != "2026-03-23 21:30" {
		t.Fatalf("unexpected draw date: %s", drawDate.Format("2006-01-02 15:04"))
	}
}

func TestResolveLocalDrawByIssueWithAnchor(t *testing.T) {
	definition := Definition{
		Code: "dlt",
		Name: "体彩大乐透",
		DrawSchedule: DrawScheduleSettings{
			Weekdays:    []int{1, 2, 6},
			Time:        "21:30",
			AnchorIssue: "2026030",
			AnchorDate:  "2026-03-23",
		},
	}

	drawDate, ok, err := resolveLocalDrawByIssue(definition, "2026029")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected local draw date")
	}
	if drawDate.Format("2006-01-02 15:04") != "2026-03-21 21:30" {
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
	issue, drawDate, err := buildRecommendationPlanFromAnchor(
		definition,
		"2026028",
		time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
		now,
	)
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
