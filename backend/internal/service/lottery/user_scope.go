package lottery

import (
	"fmt"

	userModel "go-fiber-starter/internal/model/user"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func parseRequiredUserID(userID string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(userID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("用户信息无效")
	}
	return parsed, nil
}

func loadSchedulerUsers() ([]userModel.User, error) {
	items := make([]userModel.User, 0)
	if err := db.DB.Order("created_at asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func currentUserScope(query *gorm.DB, userID string) *gorm.DB {
	if userID == "" {
		return query
	}
	return query.Where("user_id = ?", userID)
}
