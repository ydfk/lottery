package lottery

import (
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

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
		return nil
	}

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
		if !isUnfinalDrawResult(item) {
			*draw = item
			return nil
		}
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
