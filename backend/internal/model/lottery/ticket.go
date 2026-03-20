package lottery

import (
	"time"

	"go-fiber-starter/internal/model/base"

	"github.com/google/uuid"
)

type Ticket struct {
	base.BaseModel
	UserID           *uuid.UUID    `gorm:"type:uuid;index" json:"-"`
	LotteryCode      string        `gorm:"index;size:32" json:"lotteryCode"`
	RecommendationID *uuid.UUID    `gorm:"type:uuid;index" json:"recommendationId"`
	Issue            string        `gorm:"index;size:32" json:"issue"`
	EntrySignature   *string       `gorm:"size:64;index" json:"-"`
	ManualDrawDate   *time.Time    `json:"manualDrawDate"`
	Source           string        `gorm:"size:32" json:"source"`
	ImagePath        string        `gorm:"size:255" json:"imagePath"`
	RecognizedText   string        `gorm:"type:text" json:"recognizedText"`
	Status           string        `gorm:"size:32" json:"status"`
	CostAmount       float64       `json:"costAmount"`
	PrizeAmount      float64       `json:"prizeAmount"`
	PurchasedAt      time.Time     `json:"purchasedAt"`
	CheckedAt        *time.Time    `json:"checkedAt"`
	Notes            string        `gorm:"type:text" json:"notes"`
	Entries          []TicketEntry `json:"entries"`
}

type TicketEntry struct {
	base.BaseModel
	TicketID     uuid.UUID `gorm:"type:uuid;index" json:"ticketId"`
	Sequence     int       `json:"sequence"`
	RedNumbers   string    `gorm:"size:64" json:"redNumbers"`
	BlueNumbers  string    `gorm:"size:32" json:"blueNumbers"`
	Multiple     int       `json:"multiple"`
	IsAdditional bool      `json:"isAdditional"`
	IsWinning    bool      `json:"isWinning"`
	PrizeName    string    `gorm:"size:32" json:"prizeName"`
	PrizeAmount  float64   `json:"prizeAmount"`
	MatchSummary string    `gorm:"size:64" json:"matchSummary"`
}

func (Ticket) TableName() string {
	return "tickets"
}

func (TicketEntry) TableName() string {
	return "ticket_entries"
}
