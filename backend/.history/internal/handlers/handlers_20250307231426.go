package handlers

import (
"strconv"
"time"

"github.com/gofiber/fiber/v2"
"lottery-backend/internal/models"
"lottery-backend/internal/pkg/config"
"lottery-backend/internal/pkg/database"
"lottery-backend/internal/pkg/scheduler"
)

type LotteryTypeRequest struct {
Name         string `json:"name"`
ScheduleCron string `json:"schedule_cron"`
ModelName    string `json:"model_name"`
IsActive     bool   `json:"is_active"`
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() fiber.Handler {
return func(c *fiber.Ctx) error {
token := c.Get("X-Admin-Token")
if token != config.Current.Server.AdminKey {
return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
"error": "未授权访问",
})
}
return c.Next()
}
}

// CreateLotteryType 创建彩票类型
func CreateLotteryType(c *fiber.Ctx) error {
var req LotteryTypeRequest
if err := c.BodyParser(&req); err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的请求数据",
})
}

// 验证cron表达式
if err := scheduler.ValidateCron(req.ScheduleCron); err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的cron表达式",
})
}

lotteryType := models.LotteryType{
Name:         req.Name,
ScheduleCron: req.ScheduleCron,
ModelName:    req.ModelName,
IsActive:     req.IsActive,
}

// 使用事务和审计日志记录创建操作
err := database.WithAudit(0, "CREATE", "lottery_types", 0, func() error {
return database.DB.Create(&lotteryType).Error
})

if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "创建彩票类型失败",
})
}

return c.Status(fiber.StatusCreated).JSON(lotteryType)
}

// UpdateLotteryType 更新彩票类型
func UpdateLotteryType(c *fiber.Ctx) error {
id, err := strconv.ParseUint(c.Params("id"), 10, 32)
if err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的ID",
})
}

var req LotteryTypeRequest
if err := c.BodyParser(&req); err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的请求数据",
})
}

if err := scheduler.ValidateCron(req.ScheduleCron); err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的cron表达式",
})
}

err = database.WithAudit(0, "UPDATE", "lottery_types", uint(id), func() error {
return database.DB.Model(&models.LotteryType{}).Where("id = ?", id).Updates(map[string]interface{}{
"name":          req.Name,
"schedule_cron": req.ScheduleCron,
"model_name":    req.ModelName,
"is_active":     req.IsActive,
}).Error
})

if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "更新彩票类型失败",
})
}

return c.Status(fiber.StatusOK).JSON(fiber.Map{
"message": "更新成功",
})
}

// ListLotteryTypes 列出所有彩票类型
func ListLotteryTypes(c *fiber.Ctx) error {
var types []models.LotteryType
if err := database.DB.Find(&types).Error; err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "获取彩票类型列表失败",
})
}
return c.JSON(types)
}

// GetRecommendations 获取推荐记录
func GetRecommendations(c *fiber.Ctx) error {
var recommendations []models.Recommendation
query := database.DB.Order("created_at DESC")

// 处理查询参数
if lotteryType := c.Query("lottery_type"); lotteryType != "" {
query = query.Joins("JOIN lottery_types ON recommendations.lottery_type_id = lottery_types.id").
Where("lottery_types.name = ?", lotteryType)
}

if dateStr := c.Query("date"); dateStr != "" {
date, err := time.Parse("2006-01-02", dateStr)
if err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的日期格式",
})
}
query = query.Where("DATE(created_at) = DATE(?)", date)
}

if err := query.Find(&recommendations).Error; err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "获取推荐记录失败",
})
}

return c.JSON(recommendations)
}

// UpdatePurchaseStatus 更新购买状态
func UpdatePurchaseStatus(c *fiber.Ctx) error {
id, err := strconv.ParseUint(c.Params("id"), 10, 32)
if err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的ID",
})
}

var req struct {
IsPurchased bool `json:"is_purchased"`
}

if err := c.BodyParser(&req); err != nil {
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
"error": "无效的请求数据",
})
}

err = database.WithAudit(0, "UPDATE", "recommendations", uint(id), func() error {
return database.DB.Model(&models.Recommendation{}).Where("id = ?", id).
Update("is_purchased", req.IsPurchased).Error
})

if err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "更新购买状态失败",
})
}

return c.JSON(fiber.Map{
"message": "更新成功",
})
}

// GetAuditLogs 获取审计日志
func GetAuditLogs(c *fiber.Ctx) error {
var logs []models.AuditLog
query := database.DB.Order("created_at DESC")

if table := c.Query("table"); table != "" {
query = query.Where("table_name = ?", table)
}

if err := query.Find(&logs).Error; err != nil {
return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
"error": "获取审计日志失败",
})
}

return c.JSON(logs)
}
