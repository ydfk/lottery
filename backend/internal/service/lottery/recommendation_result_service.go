package lottery

import (
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"
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
	if err := db.DB.Preload("PrizeDetails").Where("lottery_code = ? AND issue = ?", code, issue).First(&draw).Error; err != nil {
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
			result := JudgeNumbers(code, entry.RedNumbers, entry.BlueNumbers, draw, prizeMap)
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
		if err := db.DB.Save(&recommendation).Error; err != nil {
			return err
		}
	}

	return nil
}
