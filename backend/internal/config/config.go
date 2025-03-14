package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AI           AIConfig            `mapstructure:"ai"`
	LotteryAPI   LotteryAPIConfig    `mapstructure:"lottery_api"`
	Database     DatabaseConfig      `mapstructure:"database"`
	Server       ServerConfig        `mapstructure:"server"`
	JWT          JWTConfig           `mapstructure:"jwt"`
	Users        []UserConfig        `mapstructure:"users"`
	LotteryTypes []LotteryTypeConfig `mapstructure:"lottery_types"`
	Scheduler    SchedulerConfig     `mapstructure:"scheduler"`

	// 阿里云OSS配置
	AliyunOSS struct {
		Endpoint        string
		AccessKeyID     string
		AccessKeySecret string
		BucketName      string
		BucketDomain    string
	}
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"`
}

type AIConfig struct {
	BaseURL      string        `mapstructure:"base_url"`
	APIKey       string        `mapstructure:"api_key"`
	Timeout      time.Duration `mapstructure:"timeout"`
	MaxRetries   int           `mapstructure:"max_retries"`
	UseProxy     bool          `mapstructure:"use_proxy"`     // 是否使用代理
	ProxyAddress string        `mapstructure:"proxy_address"` // 代理服务器地址
}

type LotteryAPIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	AppKey  string `mapstructure:"app_key"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	AdminKey string `mapstructure:"admin_key"`
}

// UserConfig 用户配置
type UserConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LotteryTypeConfig 彩票类型配置
type LotteryTypeConfig struct {
	Code         string `mapstructure:"code"`
	Name         string `mapstructure:"name"`
	ScheduleCron string `mapstructure:"schedule_cron"`
	ModelName    string `mapstructure:"model_name"`
	IsActive     bool   `mapstructure:"is_active"`
	APIEndpoint  string `mapstructure:"api_endpoint"` // 原有API端点
	CaipiaoID    int    `mapstructure:"caipiao_id"`   // 极速API的彩票ID
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	ResultFetchCron string `mapstructure:"result_fetch_cron"` // 开奖结果爬取的cron表达式
}

var Current Config

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&Current); err != nil {
		return err
	}

	return nil
}
