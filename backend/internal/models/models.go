package models

import (
	"encoding/json"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;size:50"`
	Password  string `gorm:"size:100"` // bcrypt哈希
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HashPassword 对密码进行哈希处理
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码是否正确
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// LotteryType 彩票类型配置
type LotteryType struct {
	ID           uint   `gorm:"primaryKey"`
	Code         string `gorm:"size:20;uniqueIndex"` // 彩票代码，如 fc_ssq, tc_dlt
	Name         string `gorm:"size:50"`             // 彩票名称，如 双色球, 大乐透
	ScheduleCron string `gorm:"size:20"`             // cron表达式
	ModelName    string `gorm:"size:100"`            // 对应AI模型
	IsActive     bool   `gorm:"default:true"`        // 是否启用
	APIEndpoint  string `gorm:"size:500"`            // API查询地址
	CaipiaoID    int    `gorm:"default:0"`           // 极速API的彩票ID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// AuditLog 操作审计日志
type AuditLog struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   // 操作人ID
	Action    string `gorm:"size:10"` // CREATE/UPDATE/DELETE
	TableName string `gorm:"size:50"`
	RecordID  uint
	OldData   JSON // 变更前数据
	NewData   JSON // 变更后数据
	CreatedAt time.Time
}

// Recommendation 推荐记录
type Recommendation struct {
	ID            uint      `gorm:"primaryKey"`
	LotteryTypeID uint      // 关联彩票类型
	Numbers       string    `gorm:"size:100"` // 推荐号码
	ModelName     string    `gorm:"size:100"` // 使用的模型
	DrawTime      time.Time // 开奖时间
	DrawNumber    string    `gorm:"size:20"`       // 期数，例如：23001、23002
	IsPurchased   bool      `gorm:"default:false"` // 是否已购买
	DrawResult    string    `gorm:"size:100"`      // 开奖结果
	WinStatus     string    `gorm:"size:50"`       // 中奖状态
	WinAmount     float64   // 中奖金额
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DrawResult 彩票开奖结果
type DrawResult struct {
	ID             uint      `gorm:"primaryKey"`
	LotteryTypeID  uint      // 关联彩票类型
	CaipiaoID      int       // 极速API的彩票ID
	DrawNumber     string    `gorm:"size:20"`  // 期号
	MainNumbers    string    `gorm:"size:100"` // 主号码
	SpecialNumbers string    `gorm:"size:50"`  // 特殊号码（如蓝球）
	DrawDate       time.Time // 开奖日期
	SaleAmount     float64   // 销售额
	PoolAmount     float64   // 奖池金额
	PrizeInfo      JSON      // 奖金信息，JSON格式
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
