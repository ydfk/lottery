package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"

	"lottery-backend/internal/handlers"
	"lottery-backend/internal/pkg/ai"
	"lottery-backend/internal/pkg/config"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"
	"lottery-backend/internal/pkg/middleware"
	"lottery-backend/internal/pkg/scheduler"
)

func main() {
	// 初始化日志系统
	if err := logger.Init("logs"); err != nil {
		panic(err)
	}

	// 初始化配置
	if err := config.Init(); err != nil {
		logger.Fatal("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		logger.Fatal("初始化数据库失败: %v", err)
	}

	// 创建AI客户端
	aiClient := ai.NewClient()

	// 将AI客户端传递给handlers包，使其可以在处理程序中使用
	handlers.SetAIClient(aiClient)

	//创建并启动调度器
	scheduler := scheduler.NewScheduler(aiClient)
	if err := scheduler.Start(); err != nil {
		logger.Fatal("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

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

	// API路由（需要JWT认证）
	api := app.Group("/api")

	// 认证路由（不需要JWT）
	//api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)

	// JWT 中间件（白名单Filter也可以保留，用于额外判断，但此时/login不会经过此中间件）
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(config.Current.JWT.Secret),
		Filter: func(c *fiber.Ctx) bool {
			// 如果请求的是 /api/login 则跳过JWT验证
			return c.Path() == "/api/login"
		},
	}))

	// 获取彩票类型列表
	api.Get("/lottery-types", handlers.ListLotteryTypes)

	// 创建和更新彩票类型
	api.Post("/lottery-types", handlers.CreateLotteryType)
	api.Put("/lottery-types/:id", handlers.UpdateLotteryType)

	// 推荐记录相关
	api.Get("/recommendations", handlers.GetRecommendations)
	api.Put("/recommendations/:id/purchase", handlers.UpdatePurchaseStatus)

	// 开奖历史记录
	api.Get("/draw-results", handlers.GetDrawResults)

	// 审计日志
	api.Get("/audit-logs", handlers.GetAuditLogs)

	// 添加新接口：手动触发生成彩票号码推荐
	api.Post("/lottery/generate", handlers.GenerateLotteryNumbers)

	// 添加新接口：手动触发爬取彩票开奖结果
	api.Post("/lottery/crawl", handlers.CrawlLotteryResults)

	// 启动服务器
	port := fmt.Sprintf(":%d", config.Current.Server.Port)
	if err := app.Listen(port); err != nil {
		logger.Fatal("启动服务器失败: %v", err)
	}
}
