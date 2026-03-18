package db

import (
	lotteryModel "go-fiber-starter/internal/model/lottery"
	userModel "go-fiber-starter/internal/model/user"
	"strconv"
)

func autoMigrate() error {
	if err := DB.AutoMigrate(
		&userModel.User{},
		&lotteryModel.LotteryType{},
		&lotteryModel.DrawResult{},
		&lotteryModel.DrawPrize{},
		&lotteryModel.TicketUpload{},
		&lotteryModel.Ticket{},
		&lotteryModel.TicketEntry{},
		&lotteryModel.Recommendation{},
		&lotteryModel.RecommendationEntry{},
	); err != nil {
		return err
	}

	if err := cleanupDuplicateTicketEntries(); err != nil {
		return err
	}

	return DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_ticket_entries_ticket_sequence ON ticket_entries(ticket_id, sequence)").Error
}

func cleanupDuplicateTicketEntries() error {
	entries := make([]lotteryModel.TicketEntry, 0)
	if err := DB.Order("ticket_id asc").Order("sequence asc").Order("created_at asc").Order("id asc").Find(&entries).Error; err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(entries))
	duplicateIDs := make([]any, 0)
	affectedTicketIDs := make(map[string]struct{})
	for _, entry := range entries {
		key := entry.TicketID.String() + ":" + strconv.Itoa(entry.Sequence)
		if _, exists := seen[key]; exists {
			duplicateIDs = append(duplicateIDs, entry.Id)
			affectedTicketIDs[entry.TicketID.String()] = struct{}{}
			continue
		}
		seen[key] = struct{}{}
	}

	if len(duplicateIDs) == 0 {
		return nil
	}

	if err := DB.Where("id IN ?", duplicateIDs).Delete(&lotteryModel.TicketEntry{}).Error; err != nil {
		return err
	}

	for ticketID := range affectedTicketIDs {
		if err := refreshTicketPrizeSummary(ticketID); err != nil {
			return err
		}
	}

	return nil
}

func refreshTicketPrizeSummary(ticketID string) error {
	ticket := lotteryModel.Ticket{}
	if err := DB.First(&ticket, "id = ?", ticketID).Error; err != nil {
		return err
	}

	entries := make([]lotteryModel.TicketEntry, 0)
	if err := DB.Where("ticket_id = ?", ticketID).Find(&entries).Error; err != nil {
		return err
	}

	totalPrize := 0.0
	hasWinning := false
	for _, entry := range entries {
		totalPrize += entry.PrizeAmount
		hasWinning = hasWinning || entry.IsWinning
	}

	updates := map[string]any{
		"prize_amount": totalPrize,
	}
	if ticket.CheckedAt != nil {
		if hasWinning {
			updates["status"] = "won"
		} else {
			updates["status"] = "not_won"
		}
	}

	return DB.Model(&lotteryModel.Ticket{}).Where("id = ?", ticketID).Updates(updates).Error
}
