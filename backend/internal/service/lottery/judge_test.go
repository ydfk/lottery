package lottery

import (
	"testing"

	model "go-fiber-starter/internal/model/lottery"
)

func TestJudgeDLTAdditionalPrizeAmount(t *testing.T) {
	draw := model.DrawResult{
		RedNumbers:  "03,11,18,26,32",
		BlueNumbers: "04,09",
	}

	result := JudgeNumbers("dlt", "03,11,18,26,32", "04,09", true, draw, map[string]float64{
		"一等奖": 1000,
	})
	if !result.IsWinning {
		t.Fatalf("expected winning result")
	}
	if result.PrizeAmount != 1800 {
		t.Fatalf("unexpected prize amount: %v", result.PrizeAmount)
	}
}

func TestResolveDLTPrizeNameWithNewRules(t *testing.T) {
	if prizeName := resolveDLTPrizeName(4, 2); prizeName != "三等奖" {
		t.Fatalf("unexpected prize name: %s", prizeName)
	}
	if prizeName := resolveDLTPrizeName(4, 0); prizeName != "五等奖" {
		t.Fatalf("unexpected prize name: %s", prizeName)
	}
	if prizeName := resolveDLTPrizeName(0, 2); prizeName != "七等奖" {
		t.Fatalf("unexpected prize name: %s", prizeName)
	}
}
