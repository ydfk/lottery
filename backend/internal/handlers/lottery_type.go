package handlers

import (
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"
	"lottery-backend/internal/pkg/scheduler"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type LotteryTypeRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ScheduleCron string `json:"schedule_cron"`
	ModelName    string `json:"model_name"`
	IsActive     bool   `json:"is_active"`
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

	logger.Info("成功创建彩票类型[ID:%d, Code:%s]", lotteryType.Id, lotteryType.Code)
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
