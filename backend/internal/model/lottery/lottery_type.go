package lottery

import "go-fiber-starter/internal/model/base"

type LotteryType struct {
	base.BaseModel
	Code                   string `gorm:"uniqueIndex;size:32" json:"code"`
	Name                   string `gorm:"size:64" json:"name"`
	Status                 string `gorm:"size:16" json:"status"`
	RemoteLotteryID        string `gorm:"size:32" json:"remoteLotteryId"`
	RedCount               int    `json:"redCount"`
	BlueCount              int    `json:"blueCount"`
	RedMin                 int    `json:"redMin"`
	RedMax                 int    `json:"redMax"`
	BlueMin                int    `json:"blueMin"`
	BlueMax                int    `json:"blueMax"`
	RecommendationCount    int    `json:"recommendationCount"`
	RecommendationProvider string `gorm:"size:32" json:"recommendationProvider"`
	RecommendationModel    string `gorm:"size:128" json:"recommendationModel"`
	VisionProvider         string `gorm:"size:32" json:"visionProvider"`
	VisionModel            string `gorm:"size:128" json:"visionModel"`
}

func (LotteryType) TableName() string {
	return "lottery_types"
}
