package database

import (
	"fmt"
	"lottery-backend/internal/models"
)

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.LotteryType{},
		&models.AuditLog{},
		&models.Recommendation{},
		&models.DrawResult{},
	)
}

func migrate() error {
	// 自动迁移数据表结构
	err := DB.AutoMigrate(
		&models.User{},
		&models.LotteryType{},
		&models.Recommendation{},
		&models.DrawResult{},
		&models.AuditLog{},
		&models.LotteryPurchase{}, // 添加彩票购买记录表迁移
	)

	if err != nil {
		return fmt.Errorf("自动迁移数据库失败: %v", err)
	}

	return nil
}
