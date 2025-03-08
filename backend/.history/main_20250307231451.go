package main

import (
"fmt"
"log"
"os"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/fiber/v2/middleware/cors"
"github.com/gofiber/fiber/v2/middleware/logger"
"github.com/gofiber/fiber/v2/middleware/recover"

"lottery-backend/internal/handlers"
"lottery-backend/internal/pkg/ai"
"lottery-backend/internal/pkg/config"
"lottery-backend/internal/pkg/database"
"lottery-backend/internal/pkg/scheduler"
)

func main() {
// 初始化配置
if err := config.Init(); err != nil {
log.Fatalf("加载配置失败: %v", err)
}

// 初始化数据库
if err := database.Init(); err != nil {
log.Fatalf("初始化数据库失败: %v", err)
}

// 创建AI客户端
aiClient := ai.NewClient()

// 创建并启动调度器
scheduler := scheduler.NewScheduler(aiClient)
if err := scheduler.Start(); err != nil {
log.Fatalf("启动调度器失败: %v", err)
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
app.Use(logger.New(logger.Config{
Format: "${time} ${ip} ${status} ${latency} ${method} ${path}\n",
Output: os.Stdout,
}))

// API路由
api := app.Group("/api")

// 公开API
api.Get("/lottery-types", handlers.ListLotteryTypes)
api.Get("/recommendations", handlers.GetRecommendations)

// 管理API（需要认证）
admin := api.Group("/admin", handlers.AdminAuthMiddleware())
admin.Post("/lottery-types", handlers.CreateLotteryType)
admin.Put("/lottery-types/:id", handlers.UpdateLotteryType)
admin.Put("/recommendations/:id/purchase", handlers.UpdatePurchaseStatus)
admin.Get("/audit-logs", handlers.GetAuditLogs)

// 启动服务器
port := fmt.Sprintf(":%d", config.Current.Server.Port)
if err := app.Listen(port); err != nil {
log.Fatalf("启动服务器失败: %v", err)
}
}
