package db

import (
	"fmt"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/logger"
	"go-fiber-starter/pkg/util"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

var DB *gorm.DB

func Init() error {
	path := config.Current.Database.Path
	if err := util.EnsureDir(path); err != nil {
		logger.Error("创建数据库目录失败: %w", err)
		return err
	}
	gormLogger := zapgorm2.New(logger.Logger.Desugar())
	gormLogger.IgnoreRecordNotFoundError = true

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return err
	}
	DB = db
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	return err
}
