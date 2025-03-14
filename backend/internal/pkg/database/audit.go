package database

import (
	"encoding/json"
	"lottery-backend/internal/models"
)

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
