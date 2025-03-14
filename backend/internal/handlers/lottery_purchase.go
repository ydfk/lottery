package handlers

import (
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"
	"lottery-backend/internal/pkg/storage"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreateLotteryPurchase 创建彩票购买记录
func CreateLotteryPurchase(c *fiber.Ctx) error {
	// 获取文件
	file, err := c.FormFile("image")
	if err != nil {
		logger.Error("获取上传文件失败: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "请上传彩票图片",
		})
	}

	// 读取文件内容
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("读取文件失败: %v", err),
		})
	}
	defer fileContent.Close()

	// 将文件内容读取到内存
	buffer := make([]byte, file.Size)
	if _, err := fileContent.Read(buffer); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("读取文件内容失败: %v", err),
		})
	}

	// 上传到OSS
	imageUrl, err := storage.UploadTicketImage(buffer, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("上传图片失败: %v", err),
		})
	}

	// 解析其他表单数据
	lotteryTypeID, err := strconv.ParseUint(c.FormValue("lotteryTypeId", "0"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的彩票类型ID",
		})
	}

	recommendationID, err := strconv.ParseUint(c.FormValue("recommendationId", "0"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的推荐记录ID",
		})
	}

	purchaseAmount, err := strconv.ParseFloat(c.FormValue("purchaseAmount", "0"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的购买金额",
		})
	}

	purchase := &models.LotteryPurchase{
		LotteryTypeID:    uint(lotteryTypeID),
		RecommendationID: uint(recommendationID),
		DrawNumber:       c.FormValue("drawNumber"),
		Numbers:          c.FormValue("numbers"),
		ImageUrl:         imageUrl,
		PurchaseTime:     time.Now(),
		PurchaseAmount:   purchaseAmount,
	}

	// 保存购买记录
	userID := getUserIDFromContext(c)
	err = database.WithAudit(userID, "CREATE", "lottery_purchases", 0, func() error {
		if err := database.DB.Create(purchase).Error; err != nil {
			return err
		}

		// 如果关联了推荐记录，更新推荐记录的购买状态
		if purchase.RecommendationID > 0 {
			return database.DB.Model(&models.Recommendation{}).
				Where("id = ?", purchase.RecommendationID).
				Update("is_purchased", true).Error
		}
		return nil
	})

	if err != nil {
		logger.Error("保存购买记录失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "保存购买记录失败",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(purchase)
}

// UpdateLotteryPurchase 更新彩票购买记录
func UpdateLotteryPurchase(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的记录ID",
		})
	}

	var purchase models.LotteryPurchase
	if err := database.DB.First(&purchase, uint(id)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "购买记录不存在",
		})
	}

	// 如果上传了新图片，处理图片更新
	if file, err := c.FormFile("image"); err == nil {
		// 删除旧图片
		if purchase.ImageUrl != "" {
			if err := storage.DeleteTicketImage(purchase.ImageUrl); err != nil {
				logger.Error("删除旧图片失败: %v", err)
			}
		}

		// 上传新图片
		fileContent, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("读取文件失败: %v", err),
			})
		}
		defer fileContent.Close()

		buffer := make([]byte, file.Size)
		if _, err := fileContent.Read(buffer); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("读取文件内容失败: %v", err),
			})
		}

		imageUrl, err := storage.UploadTicketImage(buffer, file.Filename)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("上传图片失败: %v", err),
			})
		}
		purchase.ImageUrl = imageUrl
	}

	// 更新其他字段
	if numbers := c.FormValue("numbers"); numbers != "" {
		purchase.Numbers = numbers
	}
	if amountStr := c.FormValue("purchaseAmount"); amountStr != "" {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "无效的购买金额",
			})
		}
		purchase.PurchaseAmount = amount
	}

	userID := getUserIDFromContext(c)
	err = database.WithAudit(userID, "UPDATE", "lottery_purchases", purchase.Id, func() error {
		return database.DB.Save(&purchase).Error
	})

	if err != nil {
		logger.Error("更新购买记录失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "更新购买记录失败",
		})
	}

	return c.JSON(purchase)
}

// DeleteLotteryPurchase 删除彩票购买记录
func DeleteLotteryPurchase(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的记录ID",
		})
	}

	var purchase models.LotteryPurchase
	if err := database.DB.First(&purchase, uint(id)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "购买记录不存在",
		})
	}

	// 删除OSS中的图片
	if purchase.ImageUrl != "" {
		if err := storage.DeleteTicketImage(purchase.ImageUrl); err != nil {
			logger.Error("删除图片失败: %v", err)
		}
	}

	userID := getUserIDFromContext(c)
	err = database.WithAudit(userID, "DELETE", "lottery_purchases", purchase.Id, func() error {
		return database.DB.Delete(&purchase).Error
	})

	if err != nil {
		logger.Error("删除购买记录失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "删除购买记录失败",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetLotteryPurchases 获取彩票购买记录列表
func GetLotteryPurchases(c *fiber.Ctx) error {
	// 解析查询参数
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	lotteryTypeID, err := strconv.ParseUint(c.Query("lotteryTypeId", "0"), 10, 32)
	if err != nil {
		lotteryTypeID = 0
	}

	drawNumber := c.Query("drawNumber")

	query := database.DB.Model(&models.LotteryPurchase{})

	// 应用过滤条件
	if lotteryTypeID > 0 {
		query = query.Where("lottery_type_id = ?", uint(lotteryTypeID))
	}
	if drawNumber != "" {
		query = query.Where("draw_number = ?", drawNumber)
	}

	// 获取总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取购买记录总数失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "获取购买记录失败",
		})
	}

	// 获取分页数据
	var purchases []models.LotteryPurchase
	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&purchases).Error; err != nil {
		logger.Error("查询购买记录失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "查询购买记录失败",
		})
	}

	return c.JSON(fiber.Map{
		"total": total,
		"data":  purchases,
		"page": fiber.Map{
			"current": page,
			"size":    pageSize,
			"total":   int(math.Ceil(float64(total) / float64(pageSize))),
		},
	})
}

// GetLotteryPurchase 获取单个彩票购买记录详情
func GetLotteryPurchase(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "无效的记录ID",
		})
	}

	var purchase models.LotteryPurchase
	if err := database.DB.First(&purchase, uint(id)).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "购买记录不存在",
		})
	}

	return c.JSON(purchase)
}
