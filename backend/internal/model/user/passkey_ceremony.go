package user

import (
	"time"

	"go-fiber-starter/internal/model/base"

	"github.com/google/uuid"
)

type PasskeyCeremony struct {
	base.BaseModel
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"-"`
	Kind        string    `gorm:"size:32;not null;index" json:"-"`
	SessionData []byte    `gorm:"type:blob;not null" json:"-"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"-"`
}
