package database

import (
	"fmt"
	"lottery-backend/internal/config"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() error {
	// 确保数据库文件目录存在
	dbPath := config.Current.Database.Path
	if err := ensureDBDir(dbPath); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	// 初始化数据库连接 - 使用modernc纯Go SQLite驱动
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        dbPath,
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	DB = db

	// 自动迁移数据库表
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 初始化用户数据
	if err := initUsers(); err != nil {
		return fmt.Errorf("初始化用户数据失败: %v", err)
	}

	// 初始化彩票类型数据
	if err := initLotteryTypes(); err != nil {
		return fmt.Errorf("初始化彩票类型数据失败: %v", err)
	}

	return nil
}

// ensureDBDir 确保数据库文件目录存在
func ensureDBDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}
