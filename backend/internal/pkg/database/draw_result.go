package database

import (
	"lottery-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// GetDrawResults 分页查询开奖结果
func GetDrawResults(db *gorm.DB, query *models.DrawResultQuery) ([]models.DrawResult, int64, error) {
	var results []models.DrawResult
	var total int64

	// 构建查询条件
	q := db.Model(&models.DrawResult{})

	// 应用过滤条件
	if query.LotteryTypeID > 0 {
		q = q.Where("lottery_type_id = ?", query.LotteryTypeID)
	}

	if query.DrawNumber != "" {
		q = q.Where("draw_number = ?", query.DrawNumber)
	}

	// 如果开始日期不为零值，添加到查询条件
	if !query.StartDate.IsZero() {
		q = q.Where("draw_date >= ?", query.StartDate)
	}

	// 如果结束日期不为零值，添加到查询条件
	if !query.EndDate.IsZero() {
		// 将结束日期设置为当天的23:59:59，以包含整天
		endDate := time.Date(query.EndDate.Year(), query.EndDate.Month(), query.EndDate.Day(), 23, 59, 59, 999999999, query.EndDate.Location())
		q = q.Where("draw_date <= ?", endDate)
	}

	// 获取总记录数
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := q.Order("draw_date DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&results).Error

	return results, total, err
}
