package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// AIClient 接口定义
var AIClient interface {
	GenerateLotteryNumbers(ctx context.Context, lotteryType string, model string) (string, error)
	GenerateMultipleLotteryNumbers(ctx context.Context, lotteryType string, model string, count int) ([]string, error)
}

// SetAIClient 设置AI客户端
func SetAIClient(client interface {
	GenerateLotteryNumbers(ctx context.Context, lotteryType string, model string) (string, error)
	GenerateMultipleLotteryNumbers(ctx context.Context, lotteryType string, model string, count int) ([]string, error)
}) {
	AIClient = client
}

// getUserIDFromContext 从JWT上下文中获取用户ID
func getUserIDFromContext(c *fiber.Ctx) uint {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	return userID
}
