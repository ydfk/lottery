package database

import (
	"encoding/json"
	"fmt"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/config"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() error {
	// 确保数据库文件目录存在
	dbPath := config.Current.Database.Path
	if err := ensureDBDir(dbPath); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	// 初始化数据库连接
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
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

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.LotteryType{},
		&models.AuditLog{},
		&models.Recommendation{},
	)
}

// CreateAuditLog 创建审计日志
func CreateAuditLog(userID uint, action string, tableName string, recordID uint, oldData, newData interface{}) error {
	var oldJSON, newJSON models.JSON

	if oldData != nil {
		data, err := json.Marshal(oldData)
		if err != nil {
			return err
		}
		oldJSON = models.JSON(data)
	}

	if newData != nil {
		data, err := json.Marshal(newData)
		if err != nil {
			return err
		}
		newJSON = models.JSON(data)
	}

	log := &models.AuditLog{
		UserID:    userID,
		Action:    action,
		TableName: tableName,
		RecordID:  recordID,
		OldData:   oldJSON,
		NewData:   newJSON,
	}

	return DB.Create(log).Error
}

// WithAudit 包装事务，自动记录审计日志
func WithAudit(userID uint, action string, tableName string, recordID uint, fn func() error) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 如果是更新操作，先获取原始数据
	var oldData interface{}
	if action == "UPDATE" || action == "DELETE" {
		switch tableName {
		case "lottery_types":
			var data models.LotteryType
			if err := tx.First(&data, recordID).Error; err != nil {
				tx.Rollback()
				return err
			}
			oldData = data
			// 可以根据需要添加其他表的处理
		}
	}

	if err := fn(); err != nil {
		tx.Rollback()
		return err
	}

	// 获取新数据（如果是更新操作）
	var newData interface{}
	if action == "UPDATE" {
		switch tableName {
		case "lottery_types":
			var data models.LotteryType
			if err := tx.First(&data, recordID).Error; err != nil {
				tx.Rollback()
				return err
			}
			newData = data
		}
	}

	// 记录审计日志
	if err := CreateAuditLog(userID, action, tableName, recordID, oldData, newData); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
