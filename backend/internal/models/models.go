package models

import (
	"encoding/json"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	Id        uint   `gorm:"primaryKey"`
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
	Id           uint   `gorm:"primaryKey"`
	Code         string `gorm:"size:20;uniqueIndex"` // 彩票代码，如 fc_ssq, tc_dlt
	Name         string `gorm:"size:50"`             // 彩票名称，如 双色球, 大乐透
	ScheduleCron string `gorm:"size:20"`             // cron表达式
	ModelName    string `gorm:"size:100"`            // 对应AI模型
	IsActive     bool   `gorm:"default:true"`        // 是否启用
	CaipiaoId    int    `gorm:"default:0"`           // 极速API的彩票ID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// AuditLog 操作审计日志
type AuditLog struct {
	Id        uint   `gorm:"primaryKey"`
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
	Id               uint       `gorm:"primaryKey"`
	LotteryTypeID    uint       // 关联彩票类型
	Numbers          string     `gorm:"size:100"` // 推荐号码
	ModelName        string     `gorm:"size:100"` // 使用的模型
	DrawTime         *time.Time // 实际开奖时间，创建推荐时可能未知
	ExpectedDrawTime time.Time  // 预计开奖时间
	DrawNumber       string     `gorm:"size:20"`       // 目标期数，下一期的期数，例如：23001、23002
	IsPurchased      bool       `gorm:"default:false"` // 是否已购买
	DrawResult       string     `gorm:"size:100"`      // 开奖结果
	WinStatus        string     `gorm:"size:50"`       // 中奖状态
	WinAmount        float64    // 中奖金额
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// DrawResult 彩票开奖结果
type DrawResult struct {
	Id             uint      `gorm:"primaryKey"`
	LotteryTypeID  uint      // 关联彩票类型
	CaipiaoID      int       // 极速API的彩票ID
	DrawNumber     string    `gorm:"size:20"`  // 期号
	MainNumbers    string    `gorm:"size:100"` // 主号码
	SpecialNumbers string    `gorm:"size:50"`  // 特殊号码（如蓝球）
	DrawDate       time.Time // 开奖日期
	SaleAmount     float64   `gorm:"default:0"` // 销售额
	PoolAmount     float64   `gorm:"default:0"` // 奖池金额

	// 奖项信息 - 普通奖项
	FirstPrize      float64 `gorm:"default:0"` // 一等奖金额
	FirstPrizeNum   int     `gorm:"default:0"` // 一等奖中奖注数
	SecondPrize     float64 `gorm:"default:0"` // 二等奖金额
	SecondPrizeNum  int     `gorm:"default:0"` // 二等奖中奖注数
	ThirdPrize      float64 `gorm:"default:0"` // 三等奖金额
	ThirdPrizeNum   int     `gorm:"default:0"` // 三等奖中奖注数
	FourthPrize     float64 `gorm:"default:0"` // 四等奖金额
	FourthPrizeNum  int     `gorm:"default:0"` // 四等奖中奖注数
	FifthPrize      float64 `gorm:"default:0"` // 五等奖金额
	FifthPrizeNum   int     `gorm:"default:0"` // 五等奖中奖注数
	SixthPrize      float64 `gorm:"default:0"` // 六等奖金额
	SixthPrizeNum   int     `gorm:"default:0"` // 六等奖中奖注数
	SeventhPrize    float64 `gorm:"default:0"` // 七等奖金额(大乐透特有)
	SeventhPrizeNum int     `gorm:"default:0"` // 七等奖中奖注数
	EighthPrize     float64 `gorm:"default:0"` // 八等奖金额(大乐透特有)
	EighthPrizeNum  int     `gorm:"default:0"` // 八等奖中奖注数
	NinthPrize      float64 `gorm:"default:0"` // 九等奖金额(大乐透特有)
	NinthPrizeNum   int     `gorm:"default:0"` // 九等奖中奖注数

	// 追加奖项(大乐透特有)
	FirstPrizeAdd      float64 `gorm:"default:0"` // 一等奖追加金额
	FirstPrizeAddNum   int     `gorm:"default:0"` // 一等奖追加中奖注数
	SecondPrizeAdd     float64 `gorm:"default:0"` // 二等奖追加金额
	SecondPrizeAddNum  int     `gorm:"default:0"` // 二等奖追加中奖注数
	ThirdPrizeAdd      float64 `gorm:"default:0"` // 三等奖追加金额
	ThirdPrizeAddNum   int     `gorm:"default:0"` // 三等奖追加中奖注数
	FourthPrizeAdd     float64 `gorm:"default:0"` // 四等奖追加金额
	FourthPrizeAddNum  int     `gorm:"default:0"` // 四等奖追加中奖注数
	FifthPrizeAdd      float64 `gorm:"default:0"` // 五等奖追加金额
	FifthPrizeAddNum   int     `gorm:"default:0"` // 五等奖追加中奖注数
	SixthPrizeAdd      float64 `gorm:"default:0"` // 六等奖追加金额
	SixthPrizeAddNum   int     `gorm:"default:0"` // 六等奖追加中奖注数
	SeventhPrizeAdd    float64 `gorm:"default:0"` // 七等奖追加金额
	SeventhPrizeAddNum int     `gorm:"default:0"` // 七等奖追加中奖注数

	// 其他信息
	OfficialOpenDate string `gorm:"size:20"` // 官方开奖日期
	Deadline         string `gorm:"size:20"` // 兑奖截止日期

	// 保存完整奖项信息(原始JSON)
	PrizeInfo JSON // 奖金信息，JSON格式，包含完整的奖项信息
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LotteryPurchase 彩票购买记录
type LotteryPurchase struct {
	Id               uint      `gorm:"primaryKey"`
	LotteryTypeID    uint      // 关联彩票类型
	RecommendationID uint      // 关联推荐记录
	DrawNumber       string    `gorm:"size:20"`  // 期号
	Numbers          string    `gorm:"size:100"` // 购买的号码
	ImageUrl         string    `gorm:"size:255"` // 彩票图片URL
	PurchaseTime     time.Time // 购买时间
	PurchaseAmount   float64   // 购买金额
	IsWin            bool      `gorm:"default:false"` // 是否中奖
	WinStatus        string    `gorm:"size:50"`       // 中奖状态
	WinAmount        float64   `gorm:"default:0"`     // 中奖金额
	CreatedAt        time.Time
	UpdatedAt        time.Time
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

// DrawResultQuery 开奖结果查询参数
type DrawResultQuery struct {
	Page          int       // 页码，从1开始
	PageSize      int       // 每页数量
	LotteryTypeID uint      // 彩票类型ID
	DrawNumber    string    // 期号
	StartDate     time.Time // 开始日期
	EndDate       time.Time // 结束日期
}
