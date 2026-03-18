package lottery

import "testing"

func TestParseDLTTextWithPackedEntries(t *testing.T) {
	text := "体彩 超级大乐透 第 26025期 2026年03月11日开奖 单式票 追加投注2倍 合计12元 ①0311182632+0409 ② 06 14 21 29 34 + 02 11"
	result, err := ParseDLTText(text)
	if err != nil {
		t.Fatalf("parse dlt text: %v", err)
	}
	if result.LotteryCode != "dlt" {
		t.Fatalf("unexpected lottery code: %s", result.LotteryCode)
	}
	if result.Issue != "26025" {
		t.Fatalf("unexpected issue: %s", result.Issue)
	}
	if result.DrawDate != "2026-03-11" {
		t.Fatalf("unexpected draw date: %s", result.DrawDate)
	}
	if result.CostAmount != 12 {
		t.Fatalf("unexpected cost amount: %v", result.CostAmount)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("unexpected entries: %d", len(result.Entries))
	}
	if !result.Entries[0].IsAdditional || !result.Entries[1].IsAdditional {
		t.Fatalf("unexpected additional flags: %+v", result.Entries)
	}
	if result.Entries[0].Multiple != 2 || result.Entries[1].Multiple != 2 {
		t.Fatalf("unexpected multiples: %+v", result.Entries)
	}
}
