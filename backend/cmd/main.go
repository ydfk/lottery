package main

import (
	"lottery-backend/internal/config"
	"lottery-backend/internal/handlers"
	"lottery-backend/internal/pkg/ai"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"
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
	api()
}
