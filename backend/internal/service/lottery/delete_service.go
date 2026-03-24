package lottery

import (
	"os"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"gorm.io/gorm"
)

func DeleteTicket(ticketID string, userID string) error {
	imagePaths := make([]string, 0, 1)

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		ticket := model.Ticket{}
		if err := currentUserScope(tx, userID).First(&ticket, "id = ?", ticketID).Error; err != nil {
			return err
		}

		if ticket.ImagePath != "" {
			imagePaths = append(imagePaths, ticket.ImagePath)
		}

		if err := tx.Where("ticket_id = ?", ticket.Id).Delete(&model.TicketEntry{}).Error; err != nil {
			return err
		}
		if ticket.ImagePath != "" {
			if err := currentUserScope(tx.Model(&model.TicketUpload{}), userID).
				Where("image_path = ?", ticket.ImagePath).
				Delete(&model.TicketUpload{}).Error; err != nil {
				return err
			}
		}
		return currentUserScope(tx, userID).Delete(&model.Ticket{}, "id = ?", ticket.Id).Error
	}); err != nil {
		return err
	}

	cleanupUnusedTicketImages(imagePaths)
	return nil
}

func DeleteRecommendation(code string, recommendationID string, userID string) error {
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		recommendation := model.Recommendation{}
		if err := currentUserScope(tx, userID).
			First(&recommendation, "id = ? AND lottery_code = ?", recommendationID, code).Error; err != nil {
			return err
		}

		if err := currentUserScope(tx.Model(&model.Ticket{}), userID).
			Where("recommendation_id = ?", recommendation.Id).
			Update("recommendation_id", nil).Error; err != nil {
			return err
		}

		if err := tx.Where("recommendation_id = ?", recommendation.Id).Delete(&model.RecommendationEntry{}).Error; err != nil {
			return err
		}
		return currentUserScope(tx, userID).Delete(&model.Recommendation{}, "id = ?", recommendation.Id).Error
	}); err != nil {
		return err
	}
	return nil
}

func cleanupUnusedTicketImages(imagePaths []string) {
	for _, imagePath := range uniqueStrings(imagePaths) {
		if imagePath == "" {
			continue
		}

		var ticketCount int64
		if err := db.DB.Model(&model.Ticket{}).Where("image_path = ?", imagePath).Count(&ticketCount).Error; err != nil || ticketCount > 0 {
			continue
		}

		var uploadCount int64
		if err := db.DB.Model(&model.TicketUpload{}).Where("image_path = ?", imagePath).Count(&uploadCount).Error; err != nil || uploadCount > 0 {
			continue
		}

		_ = os.Remove(imagePath)
	}
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
