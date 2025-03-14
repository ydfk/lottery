package database

import (
	"lottery-backend/internal/config"
	"lottery-backend/internal/models"
)

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
