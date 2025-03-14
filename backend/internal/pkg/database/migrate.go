package database

import "lottery-backend/internal/models"

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
