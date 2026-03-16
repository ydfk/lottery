package lottery

import (
	"time"

	"go-fiber-starter/internal/model/base"

	"github.com/google/uuid"
)

type DrawResult struct {
	base.BaseModel
	LotteryCode     string      `gorm:"uniqueIndex:idx_lottery_issue;size:32" json:"lotteryCode"`
	Issue           string      `gorm:"uniqueIndex:idx_lottery_issue;size:32" json:"issue"`
	DrawDate        time.Time   `json:"drawDate"`
	RedNumbers      string      `gorm:"size:64" json:"redNumbers"`
	BlueNumbers     string      `gorm:"size:32" json:"blueNumbers"`
	SaleAmount      float64     `json:"saleAmount"`
	PrizePoolAmount float64     `json:"prizePoolAmount"`
	Source          string      `gorm:"size:32" json:"source"`
	RawPayload      string      `gorm:"type:text" json:"rawPayload"`
	PrizeDetails    []DrawPrize `json:"prizeDetails"`
}

type DrawPrize struct {
	base.BaseModel
	DrawResultID uuid.UUID `gorm:"type:uuid;index" json:"drawResultId"`
	PrizeName    string    `gorm:"size:32" json:"prizeName"`
	PrizeRule    string    `gorm:"size:128" json:"prizeRule"`
	WinnerCount  int       `json:"winnerCount"`
	SingleBonus  float64   `json:"singleBonus"`
}

func (DrawResult) TableName() string {
	return "draw_results"
}

func (DrawPrize) TableName() string {
	return "draw_prizes"
}
