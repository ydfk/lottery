package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// LotteryType 彩票类型配置
type LotteryType struct {
	ID           uint           `gorm:"primaryKey"`
	Name         string         `gorm:"size:50;uniqueIndex"`  // 彩票名称
	ScheduleCron string         `gorm:"size:20"`             // cron表达式
	ModelName    string         `gorm:"size:100"`            // 对应AI模型
	IsActive     bool           `gorm:"default:true"`        // 是否启用
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// AuditLog 操作审计日志
type AuditLog struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint          // 操作人ID
	Action    string        `gorm:"size:10"` // CREATE/UPDATE/DELETE
	TableName string        `gorm:"size:50"`
	RecordID  uint
	OldData   JSON          // 变更前数据
	NewData   JSON          // 变更后数据
	CreatedAt time.Time
}

// Recommendation 推荐记录
type Recommendation struct {
	ID            uint           `gorm:"primaryKey"`
	LotteryTypeID uint          // 关联彩票类型
	Numbers       string        `gorm:"size:100"` // 推荐号码
	ModelName     string        `gorm:"size:100"` // 使用的模型
	DrawTime      time.Time     // 开奖时间
	IsPurchased   bool         `gorm:"default:false"` // 是否已购买
	DrawResult    string        `gorm:"size:100"` // 开奖结果
	WinStatus     string        `gorm:"size:50"`  // 中奖状态
	WinAmount     float64       // 中奖金额
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// JSON 自定义JSON类型
type JSON json.RawMessage

// Scan 实现 sql.Scanner 接口
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	*j = append((*j)[0:0], bytes...)
	return nil
}

// Value 实现 driver.Valuer 接口
func (j JSON) Value() (interface{}, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}
