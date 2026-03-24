package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig
	Jwt       JwtConfig
	Database  DatabaseConfig
	Storage   StorageConfig
	Jisu      JisuConfig
	AI        AIConnectionConfig
	Vision    VisionConnectionConfig
	Lotteries []LotteryConfig `mapstructure:"lotteries"`
}

type AppConfig struct {
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type JwtConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Path         string `mapstructure:"path"`
	DSN          string `mapstructure:"dsn"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	SSLMode      string `mapstructure:"sslMode"`
	TimeZone     string `mapstructure:"timeZone"`
	MaxIdleConns int    `mapstructure:"maxIdleConns"`
	MaxOpenConns int    `mapstructure:"maxOpenConns"`
}

type StorageConfig struct {
	UploadDir string `mapstructure:"uploadDir"`
}

type JisuConfig struct {
	BaseURL        string `mapstructure:"baseURL"`
	AppKey         string `mapstructure:"appKey"`
	TimeoutSeconds int    `mapstructure:"timeoutSeconds"`
}

type AIConnectionConfig struct {
	BaseURL        string `mapstructure:"baseURL"`
	APIKey         string `mapstructure:"apiKey"`
	TimeoutSeconds int    `mapstructure:"timeoutSeconds"`
}

type VisionConnectionConfig struct {
	Provider                  string `mapstructure:"provider"`
	Model                     string `mapstructure:"model"`
	Prompt                    string `mapstructure:"prompt"`
	BaseURL                   string `mapstructure:"baseURL"`
	APIKey                    string `mapstructure:"apiKey"`
	TimeoutSeconds            int    `mapstructure:"timeoutSeconds"`
	UseDocOrientationClassify bool   `mapstructure:"useDocOrientationClassify"`
	UseDocUnwarping           bool   `mapstructure:"useDocUnwarping"`
	UseChartRecognition       bool   `mapstructure:"useChartRecognition"`
}

type LotteryConfig struct {
	Code            string                      `mapstructure:"code"`
	Name            string                      `mapstructure:"name"`
	Enabled         bool                        `mapstructure:"enabled"`
	RemoteLotteryID string                      `mapstructure:"remoteLotteryId"`
	RedCount        int                         `mapstructure:"redCount"`
	BlueCount       int                         `mapstructure:"blueCount"`
	RedMin          int                         `mapstructure:"redMin"`
	RedMax          int                         `mapstructure:"redMax"`
	BlueMin         int                         `mapstructure:"blueMin"`
	BlueMax         int                         `mapstructure:"blueMax"`
	DrawSchedule    LotteryDrawScheduleConfig   `mapstructure:"drawSchedule"`
	Recommendation  LotteryRecommendationConfig `mapstructure:"recommendation"`
	Sync            LotterySyncRuleConfig       `mapstructure:"sync"`
}

type LotteryDrawScheduleConfig struct {
	Weekdays    []int  `mapstructure:"weekdays"`
	Time        string `mapstructure:"time"`
	AnchorIssue string `mapstructure:"anchorIssue"`
	AnchorDate  string `mapstructure:"anchorDate"`
}

type LotteryRecommendationConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Cron          string `mapstructure:"cron"`
	Provider      string `mapstructure:"provider"`
	Count         int    `mapstructure:"count"`
	HistoryWindow int    `mapstructure:"historyWindow"`
	Model         string `mapstructure:"model"`
	Prompt        string `mapstructure:"prompt"`
	PromptVersion string `mapstructure:"promptVersion"`
}

type LotterySyncRuleConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Cron        string `mapstructure:"cron"`
	HistorySize int    `mapstructure:"historySize"`
}

var Current Config
var IsProduction bool

func Init() error {
	viper.SetConfigFile("config/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if _, err := os.Stat("config/config.local.yaml"); err == nil {
		viper.SetConfigFile("config/config.local.yaml")
		if err := viper.MergeInConfig(); err != nil {
			return err
		}
	}

	if err := viper.Unmarshal(&Current); err != nil {
		return err
	}

	IsProduction = Current.App.Env == "production"
	return nil
}
