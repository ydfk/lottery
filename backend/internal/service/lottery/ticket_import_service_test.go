package lottery

import (
	"context"
	"fmt"
	"testing"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type importWorkbookRow struct {
	LotteryCode string
	Issue       string
	DrawDate    string
	PurchasedAt string
	CostAmount  string
	RedNumbers  string
	BlueNumbers string
	Multiple    string
	Additional  string
	Notes       string
}

func TestImportTicketsOverwriteExistingTicket(t *testing.T) {
	setupImportTicketTestDB(t)

	userID := uuid.New().String()
	firstDrawDate := "2099-01-01"
	secondDrawDate := "2099-01-02"
	firstPurchasedAt := "2026-03-20 09:30:00"
	secondPurchasedAt := "2026-03-21 10:45:00"

	firstWorkbook := buildImportWorkbook(t, []importWorkbookRow{
		{
			LotteryCode: "ssq",
			Issue:       "2026101",
			DrawDate:    firstDrawDate,
			PurchasedAt: firstPurchasedAt,
			CostAmount:  "2",
			RedNumbers:  "01 02 03 04 05 06",
			BlueNumbers: "07",
			Multiple:    "1",
			Additional:  "false",
			Notes:       "first import",
		},
	})

	firstResult, err := ImportTickets(context.Background(), ImportTicketsInput{
		UserID:   userID,
		Workbook: firstWorkbook,
	})
	if err != nil {
		t.Fatalf("first import failed: %v", err)
	}
	if firstResult.SuccessCount != 1 || firstResult.FailedCount != 0 {
		t.Fatalf("unexpected first import result: %+v", firstResult)
	}

	firstTicket := querySingleTicketByUser(t, userID)
	firstTicketID := firstTicket.Id.String()

	secondWorkbook := buildImportWorkbook(t, []importWorkbookRow{
		{
			LotteryCode: "ssq",
			Issue:       "2026101",
			DrawDate:    secondDrawDate,
			PurchasedAt: secondPurchasedAt,
			CostAmount:  "10",
			RedNumbers:  "01,02,03,04,05,06",
			BlueNumbers: "07",
			Multiple:    "1",
			Additional:  "false",
			Notes:       "second import should overwrite",
		},
	})

	secondResult, err := ImportTickets(context.Background(), ImportTicketsInput{
		UserID:   userID,
		Workbook: secondWorkbook,
	})
	if err != nil {
		t.Fatalf("second import failed: %v", err)
	}
	if secondResult.SuccessCount != 1 || secondResult.FailedCount != 0 {
		t.Fatalf("unexpected second import result: %+v", secondResult)
	}

	secondTicket := querySingleTicketByUser(t, userID)
	if secondTicket.Id.String() != firstTicketID {
		t.Fatalf("ticket should be overwritten in-place, id changed from %s to %s", firstTicketID, secondTicket.Id.String())
	}

	if secondTicket.Notes != "second import should overwrite" {
		t.Fatalf("notes not overwritten: %s", secondTicket.Notes)
	}
	if secondTicket.CostAmount != 10 {
		t.Fatalf("cost amount not overwritten: %v", secondTicket.CostAmount)
	}

	expectedPurchasedAt, err := time.ParseInLocation("2006-01-02 15:04:05", secondPurchasedAt, time.Local)
	if err != nil {
		t.Fatalf("parse expected purchased time: %v", err)
	}
	if !secondTicket.PurchasedAt.Equal(expectedPurchasedAt) {
		t.Fatalf("purchased_at not overwritten: got %s want %s", secondTicket.PurchasedAt.Format(time.RFC3339), expectedPurchasedAt.Format(time.RFC3339))
	}

	if secondTicket.ManualDrawDate == nil {
		t.Fatalf("manual draw date should be set")
	}
	expectedDrawDate, err := time.ParseInLocation("2006-01-02", secondDrawDate, time.Local)
	if err != nil {
		t.Fatalf("parse expected draw date: %v", err)
	}
	if secondTicket.ManualDrawDate.Format("2006-01-02") != expectedDrawDate.Format("2006-01-02") {
		t.Fatalf("draw date not overwritten: got %s want %s", secondTicket.ManualDrawDate.Format("2006-01-02"), expectedDrawDate.Format("2006-01-02"))
	}

	if len(secondTicket.Entries) != 1 {
		t.Fatalf("expected exactly 1 entry, got %d", len(secondTicket.Entries))
	}
}

func TestImportTicketsAutoFieldsAndCodeAlias(t *testing.T) {
	setupImportTicketTestDB(t)

	userID := uuid.New().String()
	workbook := buildImportWorkbook(t, []importWorkbookRow{
		{
			LotteryCode: "dlt",
			Issue:       "26030",
			DrawDate:    "2099/03/23",
			RedNumbers:  "0311182632",
			BlueNumbers: "0409",
			Multiple:    "2",
			Additional:  "是",
			Notes:       "auto fields",
		},
		{
			LotteryCode: "dlt",
			Issue:       "26030",
			DrawDate:    "2099/03/23",
			RedNumbers:  "0614212934",
			BlueNumbers: "0211",
			Multiple:    "1",
			Additional:  "否",
			Notes:       "same issue merge",
		},
	})

	result, err := ImportTickets(context.Background(), ImportTicketsInput{
		UserID:   userID,
		Workbook: workbook,
	})

	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("unexpected import result: %+v", result)
	}

	ticket := querySingleTicketByUser(t, userID)
	if ticket.LotteryCode != "dlt" {
		t.Fatalf("lottery code not normalized: %s", ticket.LotteryCode)
	}
	if ticket.Issue != "2026030" {
		t.Fatalf("dlt issue not normalized: %s", ticket.Issue)
	}
	if ticket.CostAmount != 8 {
		t.Fatalf("cost amount should be calculated by entries: %v", ticket.CostAmount)
	}
	if ticket.ManualDrawDate == nil || ticket.ManualDrawDate.Format("2006-01-02") != "2099-03-23" {
		t.Fatalf("slash draw date not parsed correctly: %v", ticket.ManualDrawDate)
	}
	if ticket.PurchasedAt.Format("2006-01-02") != "2099-03-23" {
		t.Fatalf("purchased_at should use draw date when empty, got %s", ticket.PurchasedAt.Format("2006-01-02"))
	}
	if len(ticket.Entries) != 2 || !ticket.Entries[0].IsAdditional || ticket.Entries[0].Multiple != 2 {
		t.Fatalf("entry additional or multiple not imported correctly: %+v", ticket.Entries)
	}
}

func TestImportTicketsOverwriteSameIssueWithImportedEntries(t *testing.T) {
	setupImportTicketTestDB(t)

	userID := uuid.New().String()
	firstWorkbook := buildImportWorkbook(t, []importWorkbookRow{
		{
			LotteryCode: "ssq",
			Issue:       "2026051",
			DrawDate:    "2026/5/7",
			RedNumbers:  "010203040506",
			BlueNumbers: "07",
			Multiple:    "1",
			Notes:       "old entry",
		},
	})
	if _, err := ImportTickets(context.Background(), ImportTicketsInput{UserID: userID, Workbook: firstWorkbook}); err != nil {
		t.Fatalf("first import failed: %v", err)
	}

	secondWorkbook := buildImportWorkbook(t, []importWorkbookRow{
		{
			LotteryCode: "ssq",
			Issue:       "2026051",
			DrawDate:    "2026/5/7",
			RedNumbers:  "020914222530",
			BlueNumbers: "10",
			Multiple:    "2",
			Notes:       "new entry one",
		},
		{
			LotteryCode: "ssq",
			Issue:       "2026051",
			DrawDate:    "2026/5/7",
			RedNumbers:  "030613182433",
			BlueNumbers: "04",
			Multiple:    "2",
			Notes:       "new entry two",
		},
	})
	result, err := ImportTickets(context.Background(), ImportTicketsInput{UserID: userID, Workbook: secondWorkbook})
	if err != nil {
		t.Fatalf("second import failed: %v", err)
	}
	if result.SuccessCount != 2 || result.FailedCount != 0 {
		t.Fatalf("unexpected second import result: %+v", result)
	}

	ticket := querySingleTicketByUser(t, userID)
	if ticket.CostAmount != 8 {
		t.Fatalf("cost amount should be overwritten by imported entries: %v", ticket.CostAmount)
	}
	if ticket.Notes != "new entry one\nnew entry two" {
		t.Fatalf("notes should be overwritten by imported rows: %q", ticket.Notes)
	}
	if len(ticket.Entries) != 2 {
		t.Fatalf("same issue should keep one ticket with 2 imported entries, got %d", len(ticket.Entries))
	}
	if ticket.Entries[0].RedNumbers != "02,09,14,22,25,30" || ticket.Entries[0].BlueNumbers != "10" {
		t.Fatalf("old entries were not replaced: %+v", ticket.Entries)
	}
	if ticket.Entries[1].RedNumbers != "03,06,13,18,24,33" || ticket.Entries[1].BlueNumbers != "04" {
		t.Fatalf("second imported entry missing: %+v", ticket.Entries)
	}
}

func TestImportTicketParsersSupportSlashDateAndCompactNumbers(t *testing.T) {
	for _, value := range []string{"2026/05/09", "2026/5/9"} {
		drawDate, err := parseImportDate(value)
		if err != nil {
			t.Fatalf("slash date should be parsed: %v", err)
		}
		if drawDate.Format("2006-01-02") != "2026-05-09" {
			t.Fatalf("unexpected slash date parse result: %s", drawDate.Format("2006-01-02"))
		}
	}
	excelDrawDate, err := parseImportDate("46151")
	if err != nil {
		t.Fatalf("excel date serial should be parsed: %v", err)
	}
	if excelDrawDate.Format("2006-01-02") != "2026-05-09" {
		t.Fatalf("unexpected excel date serial parse result: %s", excelDrawDate.Format("2006-01-02"))
	}

	cases := map[string]string{
		"0102030405":   "01,02,03,04,05",
		"010203040506": "01,02,03,04,05,06",
		"0409":         "04,09",
	}
	for input, expected := range cases {
		if actual := normalizeImportedNumbers(input); actual != expected {
			t.Fatalf("compact numbers not normalized: %s got %s want %s", input, actual, expected)
		}
	}
}

func setupImportTicketTestDB(t *testing.T) {
	t.Helper()

	prevConfig := config.Current
	t.Cleanup(func() {
		config.Current = prevConfig
	})
	config.Current.Lotteries = []config.LotteryConfig{
		{
			Code:      "ssq",
			Name:      "双色球",
			Enabled:   true,
			RedCount:  6,
			BlueCount: 1,
			RedMin:    1,
			RedMax:    33,
			BlueMin:   1,
			BlueMax:   16,
			DrawSchedule: config.LotteryDrawScheduleConfig{
				Time: "21:15",
			},
		},
		{
			Code:      "dlt",
			Name:      "体彩大乐透",
			Enabled:   true,
			RedCount:  5,
			BlueCount: 2,
			RedMin:    1,
			RedMax:    35,
			BlueMin:   1,
			BlueMax:   12,
			DrawSchedule: config.LotteryDrawScheduleConfig{
				Time: "21:15",
			},
		},
	}

	prevDB := db.DB
	t.Cleanup(func() {
		db.DB = prevDB
	})

	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := gormDB.AutoMigrate(
		&model.Ticket{},
		&model.TicketEntry{},
		&model.Recommendation{},
		&model.RecommendationEntry{},
		&model.DrawResult{},
		&model.DrawPrize{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	db.DB = gormDB
}

func querySingleTicketByUser(t *testing.T, userID string) model.Ticket {
	t.Helper()

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("parse user id: %v", err)
	}

	tickets := make([]model.Ticket, 0)
	if err := db.DB.Preload("Entries").Where("user_id = ?", parsedUserID).Find(&tickets).Error; err != nil {
		t.Fatalf("query tickets: %v", err)
	}
	if len(tickets) != 1 {
		t.Fatalf("expected only one ticket after overwrite import, got %d", len(tickets))
	}
	return tickets[0]
}

func buildImportWorkbook(t *testing.T, rows []importWorkbookRow) []byte {
	t.Helper()

	workbook := excelize.NewFile()
	sheet := workbook.GetSheetName(0)
	if err := workbook.SetSheetRow(sheet, "A1", &[]interface{}{
		"lotteryCode", "issue", "drawDate", "purchasedAt", "costAmount", "redNumbers", "blueNumbers", "multiple", "isAdditional", "notes",
	}); err != nil {
		t.Fatalf("set header row: %v", err)
	}

	for index, row := range rows {
		cell := fmt.Sprintf("A%d", index+2)
		if err := workbook.SetSheetRow(sheet, cell, &[]interface{}{
			row.LotteryCode,
			row.Issue,
			row.DrawDate,
			row.PurchasedAt,
			row.CostAmount,
			row.RedNumbers,
			row.BlueNumbers,
			row.Multiple,
			row.Additional,
			row.Notes,
		}); err != nil {
			t.Fatalf("set data row: %v", err)
		}
	}

	buffer, err := workbook.WriteToBuffer()
	if err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return buffer.Bytes()
}
