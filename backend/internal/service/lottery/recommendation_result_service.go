package lottery

import (
	"context"
	"fmt"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

func RecheckRecommendation(ctx context.Context, code string, recommendationID string, userID string) (*RecommendationDetail, error) {
	recommendation := model.Recommendation{}
	if err := currentUserScope(db.DB.Preload("Entries"), userID).First(&recommendation, "id = ? AND lottery_code = ?", recommendationID, code).Error; err != nil {
		return nil, err
	}

	issue := normalizeIssueByCode(recommendation.LotteryCode, recommendation.Issue)
	if issue == "" {
		return nil, fmt.Errorf("推荐期号不能为空")
	}
	if issue != recommendation.Issue {
		recommendation.Issue = issue
		if err := db.DB.Omit("Entries").Save(&recommendation).Error; err != nil {
			return nil, err
		}
	}

	if err := findRecommendationSettlementDraw(recommendation.LotteryCode, issue, &model.DrawResult{}); err != nil {
		if _, syncErr := SyncLatestDraw(ctx, recommendation.LotteryCode, issue); syncErr != nil {
			return buildRecommendationDetail(recommendation, userID)
		}
	}
	if err := EvaluateRecommendationsByIssue(recommendation.LotteryCode, issue); err != nil {
		return nil, err
	}
	return GetRecommendationDetail(recommendation.LotteryCode, recommendation.Id.String(), userID)
}

func EvaluateRecommendationsByIssue(code string, issue string) error {
	recommendations := make([]model.Recommendation, 0)
	if err := db.DB.Preload("Entries").Where("lottery_code = ? AND issue = ?", code, issue).Find(&recommendations).Error; err != nil {
		return err
	}
	if len(recommendations) == 0 {
		return nil
	}

	draw := model.DrawResult{}
	if err := findRecommendationSettlementDraw(code, issue, &draw); err != nil {
		if _, syncErr := SyncLatestDraw(context.Background(), code, issue); syncErr == nil {
			if retryErr := findRecommendationSettlementDraw(code, issue, &draw); retryErr == nil {
				return evaluateRecommendationsWithDraw(recommendations, code, draw)
			}
		}
		return nil
	}
	return evaluateRecommendationsWithDraw(recommendations, code, draw)
}

func evaluateRecommendationsWithDraw(recommendations []model.Recommendation, code string, draw model.DrawResult) error {
	prizeMap := make(map[string]float64, len(draw.PrizeDetails))
	for _, prize := range draw.PrizeDetails {
		prizeMap[normalizePrizeName(prize.PrizeName)] = prize.SingleBonus
	}

	checkedAt := time.Now()
	for _, recommendation := range recommendations {
		totalPrize := 0.0
		for _, entry := range recommendation.Entries {
			result := JudgeNumbers(code, entry.RedNumbers, entry.BlueNumbers, false, draw, prizeMap)
			entry.IsWinning = result.IsWinning
			entry.PrizeName = result.PrizeName
			entry.PrizeAmount = result.PrizeAmount
			entry.MatchSummary = result.MatchSummary
			totalPrize += result.PrizeAmount
			if err := db.DB.Save(&entry).Error; err != nil {
				return err
			}
		}

		recommendation.CheckedAt = &checkedAt
		recommendation.PrizeAmount = totalPrize
		if err := db.DB.Omit("Entries").Save(&recommendation).Error; err != nil {
			return err
		}
	}

	return nil
}

func findRecommendationSettlementDraw(code string, issue string, draw *model.DrawResult) error {
	items := make([]model.DrawResult, 0)
	if err := db.DB.Preload("PrizeDetails").
		Where("lottery_code = ? AND issue = ?", code, issue).
		Order("created_at desc").
		Find(&items).Error; err != nil {
		return err
	}
	for _, item := range items {
		*draw = item
		return nil
	}
	return gorm.ErrRecordNotFound
}

func resetRecommendationsPendingByIssueWithDB(tx *gorm.DB, code string, issue string) error {
	recommendations := make([]model.Recommendation, 0)
	if err := tx.Preload("Entries").Where("lottery_code = ? AND issue IN ?", code, issueAliases(code, issue)).Find(&recommendations).Error; err != nil {
		return err
	}

	for _, recommendation := range recommendations {
		for _, entry := range recommendation.Entries {
			if err := tx.Model(&model.RecommendationEntry{}).
				Where("id = ?", entry.Id).
				Updates(map[string]any{
					"is_winning":    false,
					"prize_name":    "",
					"prize_amount":  0,
					"match_summary": "待开奖",
				}).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&model.Recommendation{}).
			Where("id = ?", recommendation.Id).
			Updates(map[string]any{
				"checked_at":   nil,
				"prize_amount": 0,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}
