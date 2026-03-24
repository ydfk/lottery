package lottery

import (
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

func normalizeLocalScheduleDates(definition Definition) error {
	schedule, err := parseDrawSchedule(definition)
	if err != nil {
		return err
	}

	if err := cleanupUnfinalDraws(definition); err != nil {
		return err
	}

	if _, ok, err := parseConfiguredAnchor(definition, schedule); err != nil {
		return err
	} else if !ok {
		return nil
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := normalizeDrawResultDates(tx, definition); err != nil {
			return err
		}
		return normalizeRecommendationDates(tx, definition)
	})
}

func cleanupUnfinalDraws(definition Definition) error {
	draws := make([]model.DrawResult, 0)
	if err := db.DB.Where("lottery_code = ?", definition.Code).Find(&draws).Error; err != nil {
		return err
	}

	issues := make(map[string]struct{}, len(draws))
	for _, draw := range draws {
		if !isUnfinalDrawResult(draw) {
			continue
		}
		issues[draw.Issue] = struct{}{}
	}
	for issue := range issues {
		if err := cleanupUnfinalDrawByIssue(definition.Code, issue); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDrawResultDates(tx *gorm.DB, definition Definition) error {
	items := make([]model.DrawResult, 0)
	if err := tx.Where("lottery_code = ?", definition.Code).Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		expectedDrawDate, ok, err := resolveLocalDrawByIssue(definition, item.Issue)
		if err != nil {
			return err
		}
		if !ok || sameDateTime(item.DrawDate, expectedDrawDate) {
			continue
		}
		if err := tx.Model(&model.DrawResult{}).
			Where("id = ?", item.Id).
			Update("draw_date", expectedDrawDate).Error; err != nil {
			return err
		}
	}
	return nil
}

func normalizeRecommendationDates(tx *gorm.DB, definition Definition) error {
	items := make([]model.Recommendation, 0)
	if err := tx.Where("lottery_code = ?", definition.Code).Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		expectedDrawDate, ok, err := resolveLocalDrawByIssue(definition, item.Issue)
		if err != nil {
			return err
		}
		if !ok || (item.DrawDate != nil && sameDateTime(*item.DrawDate, expectedDrawDate)) {
			continue
		}
		if err := tx.Model(&model.Recommendation{}).
			Where("id = ?", item.Id).
			Update("draw_date", expectedDrawDate).Error; err != nil {
			return err
		}
	}
	return nil
}

func sameDateTime(left time.Time, right time.Time) bool {
	if left.IsZero() && right.IsZero() {
		return true
	}
	return left.In(time.Local).Equal(right.In(time.Local))
}
