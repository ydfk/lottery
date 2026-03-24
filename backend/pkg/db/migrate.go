package db

import (
	"crypto/sha256"
	"encoding/hex"
	lotteryModel "go-fiber-starter/internal/model/lottery"
	userModel "go-fiber-starter/internal/model/user"
	"sort"
	"strconv"
	"strings"
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
	if err := backfillLotteryUserOwnership(); err != nil {
		return err
	}
	if err := backfillTicketEntrySignatures(); err != nil {
		return err
	}
	if err := cleanupDuplicateTickets(); err != nil {
		return err
	}

	if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_ticket_entries_ticket_sequence ON ticket_entries(ticket_id, sequence)").Error; err != nil {
		return err
	}
	if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tickets_user_lottery_issue_signature ON tickets(user_id, lottery_code, issue, entry_signature)").Error; err != nil {
		return err
	}
	if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_recommendations_user_lottery_issue ON recommendations(user_id, lottery_code, issue)").Error; err != nil {
		return err
	}
	return DB.Exec("CREATE INDEX IF NOT EXISTS idx_tickets_user_lottery_issue ON tickets(user_id, lottery_code, issue)").Error
}

func backfillLotteryUserOwnership() error {
	users := make([]userModel.User, 0)
	if err := DB.Order("created_at asc").Find(&users).Error; err != nil {
		return err
	}
	if len(users) != 1 {
		return nil
	}

	userID := users[0].Id
	if err := DB.Model(&lotteryModel.TicketUpload{}).Where("user_id IS NULL").Update("user_id", userID).Error; err != nil {
		return err
	}
	if err := DB.Model(&lotteryModel.Ticket{}).Where("user_id IS NULL").Update("user_id", userID).Error; err != nil {
		return err
	}
	return DB.Model(&lotteryModel.Recommendation{}).Where("user_id IS NULL").Update("user_id", userID).Error
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

func backfillTicketEntrySignatures() error {
	tickets := make([]lotteryModel.Ticket, 0)
	if err := DB.Preload("Entries").Find(&tickets).Error; err != nil {
		return err
	}

	for _, ticket := range tickets {
		signature := buildTicketEntrySignature(ticket.Entries)
		var value any
		if signature != "" {
			value = signature
		} else {
			value = nil
		}
		if err := DB.Model(&lotteryModel.Ticket{}).Where("id = ?", ticket.Id).Update("entry_signature", value).Error; err != nil {
			return err
		}
	}
	return nil
}

func cleanupDuplicateTickets() error {
	tickets := make([]lotteryModel.Ticket, 0)
	if err := DB.Where("entry_signature IS NOT NULL AND entry_signature <> ''").Order("created_at asc").Order("id asc").Find(&tickets).Error; err != nil {
		return err
	}

	seen := make(map[string]string, len(tickets))
	duplicateIDs := make([]any, 0)
	for _, ticket := range tickets {
		userID := ""
		if ticket.UserID != nil {
			userID = ticket.UserID.String()
		}
		key := strings.Join([]string{userID, ticket.LotteryCode, ticket.Issue, dereferenceString(ticket.EntrySignature)}, ":")
		if _, exists := seen[key]; exists {
			duplicateIDs = append(duplicateIDs, ticket.Id)
			continue
		}
		seen[key] = ticket.Id.String()
	}
	if len(duplicateIDs) == 0 {
		return nil
	}
	if err := DB.Where("ticket_id IN ?", duplicateIDs).Delete(&lotteryModel.TicketEntry{}).Error; err != nil {
		return err
	}
	return DB.Where("id IN ?", duplicateIDs).Delete(&lotteryModel.Ticket{}).Error
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

func buildTicketEntrySignature(entries []lotteryModel.TicketEntry) string {
	if len(entries) == 0 {
		return ""
	}

	sort.Slice(entries, func(left int, right int) bool {
		if entries[left].Sequence == entries[right].Sequence {
			return entries[left].CreatedAt.Before(entries[right].CreatedAt)
		}
		return entries[left].Sequence < entries[right].Sequence
	})

	builder := strings.Builder{}
	for index, entry := range entries {
		if index > 0 {
			builder.WriteString("|")
		}
		builder.WriteString(strconv.Itoa(entry.Sequence))
		builder.WriteString(":")
		builder.WriteString(entry.RedNumbers)
		builder.WriteString(":")
		builder.WriteString(entry.BlueNumbers)
		builder.WriteString(":")
		builder.WriteString(strconv.Itoa(maxInt(1, entry.Multiple)))
		builder.WriteString(":")
		if entry.IsAdditional {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}
	}

	sum := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

func dereferenceString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
