package lottery

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateTicketUsesServerCalculatedCost(t *testing.T) {
	setupImportTicketTestDB(t)

	result, err := CreateTicket(context.Background(), CreateTicketInput{
		UserID:     uuid.New().String(),
		Code:       "ssq",
		Issue:      "2099001",
		DrawDate:   time.Now().AddDate(1, 0, 0),
		CostAmount: 999,
		Entries: []ParsedEntry{
			{Red: []int{1, 2, 3, 4, 5, 6}, Blue: []int{7}, Multiple: 2},
			{Red: []int{8, 9, 10, 11, 12, 13}, Blue: []int{14}, Multiple: 3},
		},
	})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if result.CostAmount != 10 {
		t.Fatalf("server cost mismatch: got %v want 10", result.CostAmount)
	}
}

func TestUpdateTicketUsesServerCalculatedCost(t *testing.T) {
	setupImportTicketTestDB(t)
	userID := uuid.New().String()

	created, err := CreateTicket(context.Background(), CreateTicketInput{
		UserID:   userID,
		Code:     "ssq",
		Issue:    "2099002",
		DrawDate: time.Now().AddDate(1, 0, 0),
		Entries: []ParsedEntry{
			{Red: []int{1, 2, 3, 4, 5, 6}, Blue: []int{7}, Multiple: 1},
		},
	})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	updated, err := UpdateTicket(context.Background(), UpdateTicketInput{
		UserID:     userID,
		TicketID:   created.Id.String(),
		Code:       "dlt",
		Issue:      "2099003",
		DrawDate:   time.Now().AddDate(1, 0, 0),
		CostAmount: 1,
		Entries: []ParsedEntry{
			{Red: []int{1, 2, 3, 4, 5}, Blue: []int{6, 7}, Multiple: 4, IsAdditional: true},
		},
	})
	if err != nil {
		t.Fatalf("update ticket: %v", err)
	}
	if updated.CostAmount != 12 {
		t.Fatalf("server cost mismatch: got %v want 12", updated.CostAmount)
	}
}

func TestCalculateEntriesCost(t *testing.T) {
	entries := []ParsedEntry{
		{Multiple: 2},
		{Multiple: 3, IsAdditional: true},
	}
	if actual := calculateEntriesCost(entries); actual != 13 {
		t.Fatalf("calculated cost mismatch: got %v want 13", actual)
	}
}
