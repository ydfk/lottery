package handlers

import (
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// GetAuditLogs 获取审计日志
func GetAuditLogs(c *fiber.Ctx) error {
	logger.Info("开始获取审计日志...")

	var logs []models.AuditLog
	query := database.DB.Order("created_at DESC")

	if table := c.Query("table"); table != "" {
		logger.Info("按表名[%s]筛选...", table)
		query = query.Where("table_name = ?", table)
	}

	if err := query.Find(&logs).Error; err != nil {
		logger.Error("获取审计日志失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取审计日志失败",
		})
	}

	logger.Info("成功获取审计日志，共%d条记录", len(logs))
	return c.JSON(logs)
}
