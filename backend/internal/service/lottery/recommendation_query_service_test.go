package lottery

import (
	"testing"
	"time"

	"go-fiber-starter/internal/model/base"
	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
)

func TestQueryRecommendationsPendingFirstIsStableAcrossPages(t *testing.T) {
	setupImportTicketTestDB(t)
	userID := uuid.New()
	otherUserID := uuid.New()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	checkedAt := now

	createQueryRecommendation(t, userID, "checked-new", now.Add(4*time.Hour), &checkedAt, 10)
	createQueryRecommendation(t, userID, "pending-old", now.Add(time.Hour), nil, 0)
	createQueryRecommendation(t, userID, "checked-old", now, &checkedAt, 0)
	createQueryRecommendation(t, userID, "pending-new", now.Add(3*time.Hour), nil, 0)
	createQueryRecommendation(t, otherUserID, "other-user", now.Add(5*time.Hour), nil, 0)

	firstPage, err := QueryRecommendations(RecommendationQueryOptions{
		UserID: userID.String(), Page: 1, PageSize: 2, Sort: "latest", PendingFirst: true,
	})
	if err != nil {
		t.Fatalf("query first page: %v", err)
	}
	assertRecommendationIssues(t, firstPage.Items, []string{"pending-new", "pending-old"})
	if !firstPage.HasMore || firstPage.Total != 4 {
		t.Fatalf("unexpected first page metadata: %+v", firstPage)
	}

	secondPage, err := QueryRecommendations(RecommendationQueryOptions{
		UserID: userID.String(), Page: 2, PageSize: 2, Sort: "latest", PendingFirst: true,
	})
	if err != nil {
		t.Fatalf("query second page: %v", err)
	}
	assertRecommendationIssues(t, secondPage.Items, []string{"checked-new", "checked-old"})
}

func TestQueryRecommendationsPendingFirstKeepsSecondarySortAndStatusFilter(t *testing.T) {
	setupImportTicketTestDB(t)
	userID := uuid.New()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	checkedAt := now

	createQueryRecommendation(t, userID, "pending-new", now.Add(3*time.Hour), nil, 0)
	createQueryRecommendation(t, userID, "pending-old", now.Add(time.Hour), nil, 0)
	createQueryRecommendation(t, userID, "won", now.Add(2*time.Hour), &checkedAt, 20)
	createQueryRecommendation(t, userID, "not-won", now, &checkedAt, 0)

	oldest, err := QueryRecommendations(RecommendationQueryOptions{
		UserID: userID.String(), Page: 1, PageSize: 10, Sort: "oldest", PendingFirst: true,
	})
	if err != nil {
		t.Fatalf("query oldest: %v", err)
	}
	assertRecommendationIssues(t, oldest.Items, []string{"pending-old", "pending-new", "not-won", "won"})

	won, err := QueryRecommendations(RecommendationQueryOptions{
		UserID: userID.String(), Page: 1, PageSize: 10, Status: TicketStatusWon,
		Sort: "latest", PendingFirst: true,
	})
	if err != nil {
		t.Fatalf("query won: %v", err)
	}
	assertRecommendationIssues(t, won.Items, []string{"won"})
}

func createQueryRecommendation(
	t *testing.T,
	userID uuid.UUID,
	issue string,
	createdAt time.Time,
	checkedAt *time.Time,
	prizeAmount float64,
) {
	t.Helper()
	recommendation := model.Recommendation{
		BaseModel:   base.BaseModel{CreatedAt: createdAt, UpdatedAt: createdAt},
		UserID:      &userID,
		LotteryCode: "ssq",
		Issue:       issue,
		CheckedAt:   checkedAt,
		PrizeAmount: prizeAmount,
	}
	if err := db.DB.Create(&recommendation).Error; err != nil {
		t.Fatalf("create recommendation: %v", err)
	}
}

func assertRecommendationIssues(t *testing.T, items []RecommendationDetail, expected []string) {
	t.Helper()
	actual := make([]string, 0, len(items))
	for _, item := range items {
		actual = append(actual, item.Issue)
	}
	if len(actual) != len(expected) {
		t.Fatalf("unexpected item count: got=%v want=%v", actual, expected)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("unexpected issues: got=%v want=%v", actual, expected)
		}
	}
}
