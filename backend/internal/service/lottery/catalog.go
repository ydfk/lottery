package lottery

import (
	"errors"
	"fmt"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

const (
	ProviderPaddleOCR        = "paddleocr"
	ProviderOpenAICompatible = "openai-compatible"

	TicketStatusPending = "pending"
	TicketStatusWon     = "won"
	TicketStatusNotWon  = "not_won"
)

type RecommendationSettings struct {
	Enabled       bool
	Cron          string
	Count         int
	HistoryWindow int
	Model         string
	Prompt        string
	PromptVersion string
}

type DrawScheduleSettings struct {
	Weekdays []int
	Time     string
}

type SyncSettings struct {
	Enabled     bool
	HistorySize int
	Cron        string
}

type Definition struct {
	Code            string
	Name            string
	Enabled         bool
	RemoteLotteryID string
	RedCount        int
	BlueCount       int
	RedMin          int
	RedMax          int
	BlueMin         int
	BlueMax         int
	DrawSchedule    DrawScheduleSettings
	Recommendation  RecommendationSettings
	Sync            SyncSettings
}

func ListDefinitions() []Definition {
	definitions := make([]Definition, 0, len(config.Current.Lotteries))
	for _, item := range config.Current.Lotteries {
		definitions = append(definitions, Definition{
			Code:            item.Code,
			Name:            item.Name,
			Enabled:         item.Enabled,
			RemoteLotteryID: item.RemoteLotteryID,
			RedCount:        item.RedCount,
			BlueCount:       item.BlueCount,
			RedMin:          item.RedMin,
			RedMax:          item.RedMax,
			BlueMin:         item.BlueMin,
			BlueMax:         item.BlueMax,
			DrawSchedule: DrawScheduleSettings{
				Weekdays: append([]int(nil), item.DrawSchedule.Weekdays...),
				Time:     item.DrawSchedule.Time,
			},
			Recommendation: RecommendationSettings{
				Enabled:       item.Recommendation.Enabled,
				Cron:          item.Recommendation.Cron,
				Count:         item.Recommendation.Count,
				HistoryWindow: item.Recommendation.HistoryWindow,
				Model:         item.Recommendation.Model,
				Prompt:        item.Recommendation.Prompt,
				PromptVersion: item.Recommendation.PromptVersion,
			},
			Sync: SyncSettings{
				Enabled:     item.Sync.Enabled,
				HistorySize: item.Sync.HistorySize,
				Cron:        item.Sync.Cron,
			},
		})
	}
	return definitions
}

func GetDefinition(code string) (Definition, error) {
	for _, definition := range ListDefinitions() {
		if definition.Code == code {
			return definition, nil
		}
	}
	return Definition{}, fmt.Errorf("未找到彩种配置: %s", code)
}

func SeedLotteryTypes() error {
	definitions := ListDefinitions()
	if len(definitions) == 0 {
		return fmt.Errorf("配置文件中未找到任何彩票配置")
	}

	for _, definition := range definitions {
		item := model.LotteryType{}
		err := db.DB.Where("code = ?", definition.Code).First(&item).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item = model.LotteryType{
				Code:                   definition.Code,
				Name:                   definition.Name,
				Status:                 statusFromEnabled(definition.Enabled),
				RemoteLotteryID:        definition.RemoteLotteryID,
				RedCount:               definition.RedCount,
				BlueCount:              definition.BlueCount,
				RedMin:                 definition.RedMin,
				RedMax:                 definition.RedMax,
				BlueMin:                definition.BlueMin,
				BlueMax:                definition.BlueMax,
				RecommendationCount:    max(1, definition.Recommendation.Count),
				RecommendationProvider: ProviderOpenAICompatible,
				RecommendationModel:    definition.Recommendation.Model,
				VisionProvider:         resolveValue(config.Current.Vision.Provider, ProviderPaddleOCR),
				VisionModel:            resolveValue(config.Current.Vision.Model, "paddleocr"),
			}
			if createErr := db.DB.Create(&item).Error; createErr != nil {
				return createErr
			}
			continue
		}
		if err != nil {
			return err
		}

		item.Name = definition.Name
		item.Status = statusFromEnabled(definition.Enabled)
		item.RemoteLotteryID = definition.RemoteLotteryID
		item.RedCount = definition.RedCount
		item.BlueCount = definition.BlueCount
		item.RedMin = definition.RedMin
		item.RedMax = definition.RedMax
		item.BlueMin = definition.BlueMin
		item.BlueMax = definition.BlueMax
		item.RecommendationCount = max(1, definition.Recommendation.Count)
		item.RecommendationProvider = ProviderOpenAICompatible
		item.RecommendationModel = definition.Recommendation.Model
		item.VisionProvider = resolveValue(config.Current.Vision.Provider, ProviderPaddleOCR)
		item.VisionModel = resolveValue(config.Current.Vision.Model, "paddleocr")
		if saveErr := db.DB.Save(&item).Error; saveErr != nil {
			return saveErr
		}
	}
	return nil
}

func statusFromEnabled(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func resolveValue(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
