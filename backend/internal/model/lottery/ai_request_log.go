package lottery

import (
	"time"

	"github.com/google/uuid"
)

type AIRequestLogSummary struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	StartedAt   time.Time `json:"startedAt"`
	Purpose     string    `json:"purpose"`
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
	LotteryCode string    `json:"lotteryCode"`
	Issue       string    `json:"issue"`
	Attempt     int       `json:"attempt"`
	StatusCode  int       `json:"statusCode"`
	DurationMs  int64     `json:"durationMs"`
	Success     bool      `json:"success"`
}

type AIRequestLog struct {
	AIRequestLogSummary
	UserID      uuid.UUID `json:"-" gorm:"type:uuid;not null;index"`
	Method      string    `json:"method"`
	Endpoint    string    `json:"endpoint"`
	ContentType string    `json:"contentType"`
	Request     string    `json:"request" gorm:"type:text"`
	Response    string    `json:"response" gorm:"type:text"`
	Error       string    `json:"error" gorm:"type:text"`
}
