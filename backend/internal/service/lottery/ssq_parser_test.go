package lottery

import "testing"

func TestParseSSQTextWithMultiple(t *testing.T) {
	result, err := ParseSSQText("双色球 A. 03 09 14 21 25 32 - 07 (2)")
	if err != nil {
		t.Fatalf("parse ssq text: %v", err)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("unexpected entries: %d", len(result.Entries))
	}
	if result.Entries[0].Multiple != 2 {
		t.Fatalf("unexpected multiple: %d", result.Entries[0].Multiple)
	}
}

func TestParseSSQTextWithMultipleLineEntries(t *testing.T) {
	result, err := ParseSSQText("A. 03 09 14 21 25 32 - 07 (2)\nB. 06 11 17 24 29 33 - 12 (1)")
	if err != nil {
		t.Fatalf("parse ssq text: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("unexpected entries: %d", len(result.Entries))
	}
	if result.Entries[0].Multiple != 2 || result.Entries[1].Multiple != 1 {
		t.Fatalf("unexpected multiples: %+v", result.Entries)
	}
}
