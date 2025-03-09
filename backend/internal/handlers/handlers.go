package handlers

import (
	"context"
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"
	"lottery-backend/internal/pkg/scheduler"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type LotteryTypeRequest struct {
	Code         string `json:"code"`          // 彩票代码，如 fc_ssq, tc_dlt
	Name         string `json:"name"`          // 显示名称，如 双色球, 大乐透
	ScheduleCron string `json:"schedule_cron"` // cron表达式
	ModelName    string `json:"model_name"`    // AI模型名称
	IsActive     bool   `json:"is_active"`     // 是否启用
}

// getUserIDFromContext 从JWT上下文中获取用户ID
func getUserIDFromContext(c *fiber.Ctx) uint {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	return userID
}

// CreateLotteryType 创建彩票类型
func CreateLotteryType(c *fiber.Ctx) error {
	logger.Info("开始处理创建彩票类型请求...")

	var req LotteryTypeRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("请求数据解析失败: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的请求数据",
		})
	}
	logger.Info("收到创建彩票类型请求: %+v", req)

	// 验证code格式
	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "彩票代码不能为空",
		})
	}

	// 验证cron表达式
	if err := scheduler.ValidateCron(req.ScheduleCron); err != nil {
		logger.Error("无效的cron表达式[%s]: %v", req.ScheduleCron, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的cron表达式",
		})
	}

	lotteryType := models.LotteryType{
		Code:         req.Code,
		Name:         req.Name,
		ScheduleCron: req.ScheduleCron,
		ModelName:    req.ModelName,
		IsActive:     req.IsActive,
	}

	// 使用事务和审计日志记录创建操作
	userID := getUserIDFromContext(c)
	logger.Info("用户[%d]正在创建彩票类型...", userID)

	err := database.WithAudit(userID, "CREATE", "lottery_types", 0, func() error {
		return database.DB.Create(&lotteryType).Error
	})

	if err != nil {
		logger.Error("创建彩票类型失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "创建彩票类型失败",
		})
	}

	logger.Info("成功创建彩票类型[ID:%d, Code:%s]", lotteryType.ID, lotteryType.Code)
	return c.Status(fiber.StatusCreated).JSON(lotteryType)
}

// UpdateLotteryType 更新彩票类型
func UpdateLotteryType(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		logger.Error("无效的ID参数: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的ID",
		})
	}
	logger.Info("开始处理更新彩票类型请求[ID:%d]...", id)

	var req LotteryTypeRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("请求数据解析失败: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的请求数据",
		})
	}
	logger.Info("收到更新请求数据: %+v", req)

	if err := scheduler.ValidateCron(req.ScheduleCron); err != nil {
		logger.Error("无效的cron表达式[%s]: %v", req.ScheduleCron, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的cron表达式",
		})
	}

	userID := getUserIDFromContext(c)
	logger.Info("用户[%d]正在更新彩票类型[ID:%d]...", userID, id)

	err = database.WithAudit(userID, "UPDATE", "lottery_types", uint(id), func() error {
		return database.DB.Model(&models.LotteryType{}).Where("id = ?", id).Updates(map[string]interface{}{
			"code":          req.Code,
			"name":          req.Name,
			"schedule_cron": req.ScheduleCron,
			"model_name":    req.ModelName,
			"is_active":     req.IsActive,
		}).Error
	})

	if err != nil {
		logger.Error("更新彩票类型[ID:%d]失败: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "更新彩票类型失败",
		})
	}

	logger.Info("成功更新彩票类型[ID:%d]", id)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "更新成功",
	})
}

// ListLotteryTypes 列出所有彩票类型
func ListLotteryTypes(c *fiber.Ctx) error {
	logger.Info("开始获取彩票类型列表...")

	var types []models.LotteryType
	if err := database.DB.Find(&types).Error; err != nil {
		logger.Error("获取彩票类型列表失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取彩票类型列表失败",
		})
	}

	logger.Info("成功获取彩票类型列表，共%d条记录", len(types))
	return c.JSON(types)
}

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

// GenerateLotteryNumbers 手动触发生成彩票号码推荐
func GenerateLotteryNumbers(c *fiber.Ctx) error {
	typeID, err := strconv.ParseUint(c.Params("typeId"), 10, 32)
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

	// 使用code生成号码
	numbers, err := AIClient.GenerateLotteryNumbers(ctx, lotteryType.Code, lotteryType.ModelName)
	if err != nil {
		logger.Error("生成号码失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("生成号码失败: %v", err),
		})
	}
	logger.Info("成功生成号码: %s", numbers)

	// 计算开奖时间
	drawTime := time.Now().Add(24 * time.Hour)
	logger.Info("设置开奖时间为: %v", drawTime)

	// 保存推荐记录
	recommendation := models.Recommendation{
		LotteryTypeID: lotteryType.ID,
		Numbers:       numbers,
		ModelName:     lotteryType.ModelName,
		DrawTime:      drawTime,
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

	logger.Info("成功保存手动生成的推荐号码[ID:%d]", recommendation.ID)
	return c.Status(fiber.StatusCreated).JSON(recommendation)
}

// CrawlLotteryResults 手动触发彩票开奖结果爬取
func CrawlLotteryResults(c *fiber.Ctx) error {
	logger.Info("开始手动爬取开奖结果...")
	userID := getUserIDFromContext(c)
	logger.Info("触发用户ID: %d", userID)

	// 记录操作日志
	err := database.WithAudit(userID, "MANUAL_CRAWL", "lottery_results", 0, func() error {
		logger.Info("开始爬取开奖结果...")
		// TODO: 实现实际的爬取逻辑
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
		"message": "已开始爬取开奖结果",
	})
}

// 下面是包级变量，用于存储main.go中创建的对象
var AIClient interface {
	GenerateLotteryNumbers(ctx context.Context, lotteryType string, model string) (string, error)
}

// SetAIClient 设置AI客户端
func SetAIClient(client interface {
	GenerateLotteryNumbers(ctx context.Context, lotteryType string, model string) (string, error)
}) {
	AIClient = client
}
