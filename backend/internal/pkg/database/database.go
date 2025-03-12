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

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.LotteryType{},
		&models.AuditLog{},
		&models.Recommendation{},
		&models.DrawResult{}, // 添加新的开奖结果表
	)
}

// initUsers 初始化用户数据
func initUsers() error {
	// 如果没有配置用户，则跳过
	if len(config.Current.Users) == 0 {
		return nil
	}

	for _, userConfig := range config.Current.Users {
		// 检查用户是否已存在
		var count int64
		DB.Model(&models.User{}).Where("username = ?", userConfig.Username).Count(&count)
		if count > 0 {
			// 用户已存在，跳过
			continue
		}

		// 创建新用户
		user := models.User{
			Username: userConfig.Username,
			Password: userConfig.Password,
		}

		// 密码哈希处理
		if err := user.HashPassword(); err != nil {
			return err
		}

		// 保存用户
		if err := DB.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}

// initLotteryTypes 初始化彩票类型数据
func initLotteryTypes() error {
	// 如果没有配置彩票类型，则跳过
	if len(config.Current.LotteryTypes) == 0 {
		return nil
	}

	for _, typeConfig := range config.Current.LotteryTypes {
		// 检查彩票类型是否已存在（使用 code 字段）
		var count int64
		DB.Model(&models.LotteryType{}).Where("code = ?", typeConfig.Code).Count(&count)
		if count > 0 {
			// 彩票类型已存在，则更新
			var lotteryType models.LotteryType
			if err := DB.Where("code = ?", typeConfig.Code).First(&lotteryType).Error; err != nil {
				return err
			}

			lotteryType.Name = typeConfig.Name
			lotteryType.ScheduleCron = typeConfig.ScheduleCron
			lotteryType.ModelName = typeConfig.ModelName
			lotteryType.IsActive = typeConfig.IsActive
			lotteryType.CaipiaoId = typeConfig.CaipiaoID // 添加彩票ID

			if err := DB.Save(&lotteryType).Error; err != nil {
				return err
			}
			continue
		}

		// 创建新彩票类型
		lotteryType := models.LotteryType{
			Code:         typeConfig.Code,
			Name:         typeConfig.Name,
			ScheduleCron: typeConfig.ScheduleCron,
			ModelName:    typeConfig.ModelName,
			IsActive:     typeConfig.IsActive,
			CaipiaoId:    typeConfig.CaipiaoID, // 添加彩票ID
		}

		// 保存彩票类型
		if err := DB.Create(&lotteryType).Error; err != nil {
			return err
		}
	}

	return nil
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
