package database

import (
	"lottery-backend/internal/config"
	"lottery-backend/internal/models"
)

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
