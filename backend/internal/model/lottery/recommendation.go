package lottery

import (
	"time"

	"go-fiber-starter/internal/model/base"

	"github.com/google/uuid"
)

type Recommendation struct {
	base.BaseModel
	LotteryCode   string                `gorm:"index;size:32" json:"lotteryCode"`
	Issue         string                `gorm:"index;size:32" json:"issue"`
	Provider      string                `gorm:"size:32" json:"provider"`
	Model         string                `gorm:"size:128" json:"model"`
	Strategy      string                `gorm:"size:64" json:"strategy"`
	PromptVersion string                `gorm:"size:64" json:"promptVersion"`
	Summary       string                `gorm:"type:text" json:"summary"`
	Basis         string                `gorm:"type:text" json:"basis"`
	RawPayload    string                `gorm:"type:text" json:"rawPayload"`
	CheckedAt     *time.Time            `json:"checkedAt"`
	PrizeAmount   float64               `json:"prizeAmount"`
	Entries       []RecommendationEntry `json:"entries"`
}

type RecommendationEntry struct {
	base.BaseModel
	RecommendationID uuid.UUID `gorm:"type:uuid;index" json:"recommendationId"`
	Sequence         int       `json:"sequence"`
	RedNumbers       string    `gorm:"size:64" json:"redNumbers"`
	BlueNumbers      string    `gorm:"size:32" json:"blueNumbers"`
	Confidence       float64   `json:"confidence"`
	Reason           string    `gorm:"type:text" json:"reason"`
	IsWinning        bool      `json:"isWinning"`
	PrizeName        string    `gorm:"size:32" json:"prizeName"`
	PrizeAmount      float64   `json:"prizeAmount"`
	MatchSummary     string    `gorm:"size:64" json:"matchSummary"`
}

func (Recommendation) TableName() string {
	return "recommendations"
}

func (RecommendationEntry) TableName() string {
	return "recommendation_entries"
}
