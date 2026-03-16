// @title Go Fiber API
// @version 1.0
// @description Go Fiber Starter API
// @host localhost:25610
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	_ "go-fiber-starter/docs"
	lotteryService "go-fiber-starter/internal/service/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/logger"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}

	if err := config.Init(); err != nil {
		logger.Fatal("加载配置失败: %v", err)
	}

	if err := db.Init(); err != nil {
		logger.Fatal("初始化数据库失败: %v", err)
	}

	if err := lotteryService.Bootstrap(); err != nil {
		logger.Fatal("初始化彩票模块失败: %v", err)
	}

	api()
}
