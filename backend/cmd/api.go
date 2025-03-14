package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"

	"lottery-backend/internal/config"
	"lottery-backend/internal/handlers"
	"lottery-backend/internal/pkg/logger"
	"lottery-backend/internal/pkg/middleware"
)

func api() {
	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// 中间件
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format: "${time} ${ip} ${status} ${latency} ${method} ${path}\n",
		Output: os.Stdout,
	}))

	// 添加小驼峰响应转换中间件
	app.Use(middleware.CamelCaseResponse())

	// 配置路由组
	api := app.Group("/api")

	// 认证相关路由
	api.Post("/login", handlers.Login)
	//api.Post("/refresh", handlers.RefreshToken)

	api.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(config.Current.JWT.Secret),
		Filter: func(c *fiber.Ctx) bool {
			// 如果请求的是 /api/login 则跳过JWT验证
			return c.Path() == "/api/login"
		},
	}))

	// 彩票类型管理
	api.Post("/lottery-types", handlers.CreateLotteryType)
	api.Get("/lottery-types", handlers.ListLotteryTypes)
	api.Put("/lottery-types/:id", handlers.UpdateLotteryType)

	// 推荐号码管理
	api.Get("/recommendations", handlers.GetRecommendations)
	api.Put("/recommendations/:id/purchase-status", handlers.UpdatePurchaseStatus)
	api.Post("/generate", handlers.GenerateLotteryNumbers)

	// 开奖结果管理
	api.Get("/draw-results", handlers.GetDrawResults)
	api.Post("/lottery/crawl", handlers.CrawlLotteryResults)
	// 添加新接口：手动触发生成彩票号码推荐
	api.Post("/lottery/generate", handlers.GenerateLotteryNumbers)

	// 购买记录管理
	api.Post("/lottery-purchases", handlers.CreateLotteryPurchase)
	api.Get("/lottery-purchases", handlers.GetLotteryPurchases)
	api.Get("/lottery-purchases/:id", handlers.GetLotteryPurchase)
	api.Put("/lottery-purchases/:id", handlers.UpdateLotteryPurchase)
	api.Delete("/lottery-purchases/:id", handlers.DeleteLotteryPurchase)

	// 启动服务器
	port := fmt.Sprintf(":%d", config.Current.Server.Port)
	if err := app.Listen(port); err != nil {
		logger.Fatal("启动服务器失败: %v", err)
	}
}
