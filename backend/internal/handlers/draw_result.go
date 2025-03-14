package handlers

import (
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/draw"
	"lottery-backend/internal/pkg/logger"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetDrawResults 获取开奖结果历史记录
func GetDrawResults(c *fiber.Ctx) error {
	// 解析查询参数
	query := &models.DrawResultQuery{
		Page:          c.QueryInt("page", 1),
		PageSize:      c.QueryInt("pageSize", 10),
		LotteryTypeID: uint(c.QueryInt("lotteryTypeId", 0)),
		DrawNumber:    c.Query("drawNumber"),
	}

	// 解析日期参数
	if startDateStr := c.Query("startDate"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "起始日期格式错误，应为YYYY-MM-DD",
			})
		}
		query.StartDate = startDate
	}

	if endDateStr := c.Query("endDate"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "结束日期格式错误，应为YYYY-MM-DD",
			})
		}
		query.EndDate = endDate
	}

	// 查询数据
	results, total, err := database.GetDrawResults(database.DB, query)
	if err != nil {
		logger.Error("查询开奖结果失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("查询开奖结果失败: %v", err),
		})
	}

	// 返回结果
	return c.JSON(fiber.Map{
		"total": total,
		"data":  results,
		"page": fiber.Map{
			"current": query.Page,
			"size":    query.PageSize,
			"total":   int(math.Ceil(float64(total) / float64(query.PageSize))),
		},
	})
}

// CrawlLotteryResults 手动触发彩票开奖结果爬取
func CrawlLotteryResults(c *fiber.Ctx) error {
	logger.Info("开始手动爬取开奖结果...")
	userID := getUserIDFromContext(c)
	logger.Info("触发用户ID: %d", userID)

	// 记录操作日志
	err := database.WithAudit(userID, "MANUAL_CRAWL", "lottery_results", 0, func() error {
		logger.Info("开始爬取开奖结果...")
		// 调用爬取函数
		if err := draw.FetchAllActiveLotteryDrawResults(); err != nil {
			logger.Error("爬取开奖结果失败: %v", err)
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error("爬取开奖结果失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("爬取开奖结果失败: %v", err),
		})
	}

	logger.Info("成功触发开奖结果爬取")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "已成功爬取开奖结果并分析中奖情况",
	})
}
