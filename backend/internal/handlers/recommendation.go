package handlers

import (
	"context"
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/draw"
	"lottery-backend/internal/pkg/logger"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetRecommendations 获取推荐记录
func GetRecommendations(c *fiber.Ctx) error {
	logger.Info("开始获取推荐记录...")

	var recommendations []models.Recommendation
	query := database.DB.Order("created_at DESC")

	// 处理查询参数
	if lotteryCode := c.Query("code"); lotteryCode != "" {
		logger.Info("按彩票代码[%s]筛选...", lotteryCode)
		query = query.Joins("JOIN lottery_types ON recommendations.lottery_type_id = lottery_types.id").
			Where("lottery_types.code = ?", lotteryCode)
	}

	if dateStr := c.Query("date"); dateStr != "" {
		logger.Info("按日期[%s]筛选...", dateStr)
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			logger.Error("无效的日期格式[%s]: %v", dateStr, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "无效的日期格式",
			})
		}
		query = query.Where("DATE(created_at) = DATE(?)", date)
	}

	if drawNumber := c.Query("draw_number"); drawNumber != "" {
		logger.Info("按期号[%s]筛选...", drawNumber)
		query = query.Where("draw_number = ?", drawNumber)
	}

	if err := query.Find(&recommendations).Error; err != nil {
		logger.Error("获取推荐记录失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取推荐记录失败",
		})
	}

	logger.Info("成功获取推荐记录，共%d条记录", len(recommendations))
	return c.JSON(recommendations)
}

// UpdatePurchaseStatus 更新购买状态
func UpdatePurchaseStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		logger.Error("无效的ID参数: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的ID",
		})
	}
	logger.Info("开始更新推荐记录[ID:%d]的购买状态...", id)

	var req struct {
		IsPurchased bool `json:"is_purchased"`
	}

	if err := c.BodyParser(&req); err != nil {
		logger.Error("请求数据解析失败: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的请求数据",
		})
	}
	logger.Info("收到购买状态更新请求: purchased=%v", req.IsPurchased)

	userID := getUserIDFromContext(c)
	logger.Info("用户[%d]正在更新推荐记录[ID:%d]的购买状态...", userID, id)

	err = database.WithAudit(userID, "UPDATE", "recommendations", uint(id), func() error {
		return database.DB.Model(&models.Recommendation{}).Where("id = ?", id).
			Update("is_purchased", req.IsPurchased).Error
	})

	if err != nil {
		logger.Error("更新购买状态失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "更新购买状态失败",
		})
	}

	logger.Info("成功更新推荐记录[ID:%d]的购买状态为: %v", id, req.IsPurchased)
	return c.JSON(fiber.Map{
		"message": "更新成功",
	})
}

// GenerateLotteryNumbers 手动触发生成彩票号码推荐
func GenerateLotteryNumbers(c *fiber.Ctx) error {
	typeID, err := strconv.ParseUint(c.Query("typeId"), 10, 32)
	if err != nil {
		logger.Error("无效的彩票类型ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的彩票类型ID",
		})
	}
	logger.Info("开始手动生成彩票号码[类型ID:%d]...", typeID)

	// 获取彩票类型信息
	var lotteryType models.LotteryType
	if err := database.DB.First(&lotteryType, typeID).Error; err != nil {
		logger.Error("彩票类型[ID:%d]不存在: %v", typeID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "彩票类型不存在",
		})
	}
	logger.Info("找到彩票类型: %s (Code: %s)", lotteryType.Name, lotteryType.Code)

	if AIClient == nil {
		logger.Error("错误：AI客户端未初始化")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "AI客户端未初始化",
		})
	}

	ctx := context.Background()
	logger.Info("开始调用AI生成%s号码...", lotteryType.Code)

	// 获取开奖信息（日期和期号）
	drawInfo, err := draw.GetLotteryDrawInfo(lotteryType.Code, lotteryType.ScheduleCron)
	if err != nil {
		logger.Error("获取开奖信息失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("获取开奖信息失败: %v", err),
		})
	}
	logger.Info("获取到开奖信息: 日期=%v, 期号=%s", drawInfo.CurrentDrawDate, drawInfo.CurrentDrawNum)

	// 使用code生成号码
	numbers, err := AIClient.GenerateLotteryNumbers(ctx, lotteryType.Code, lotteryType.ModelName)
	if err != nil {
		logger.Error("生成号码失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("生成号码失败: %v", err),
		})
	}
	logger.Info("成功生成号码: %s", numbers)

	// 保存推荐记录
	recommendation := models.Recommendation{
		LotteryTypeID:    lotteryType.Id,
		Numbers:          numbers,
		ModelName:        lotteryType.ModelName,
		ExpectedDrawTime: drawInfo.NextDrawDate,
		DrawNumber:       drawInfo.NextDrawNum,
	}

	userID := getUserIDFromContext(c)
	logger.Info("用户[%d]手动生成的号码，正在保存...", userID)

	err = database.WithAudit(userID, "MANUAL_GENERATE", "recommendations", 0, func() error {
		return database.DB.Create(&recommendation).Error
	})

	if err != nil {
		logger.Error("保存推荐号码失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("保存推荐号码失败: %v", err),
		})
	}

	logger.Info("成功保存手动生成的推荐号码[ID:%d, 期号:%s]", recommendation.Id, recommendation.DrawNumber)
	return c.Status(fiber.StatusCreated).JSON(recommendation)
}
