package lottery

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/db"
	"go-fiber-starter/pkg/util"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var regexpMultipleAtEnd = regexp.MustCompile(`[（(]\s*(\d+)\s*[)）]\s*$`)

type ImportTicketsInput struct {
	UserID        string
	Workbook      []byte
	ImagesArchive []byte
}

type TicketImportRowResult struct {
	Row         int    `json:"row"`
	LotteryCode string `json:"lotteryCode,omitempty"`
	Issue       string `json:"issue,omitempty"`
	TicketID    string `json:"ticketId,omitempty"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

type TicketImportResult struct {
	TotalCount   int                     `json:"totalCount"`
	SuccessCount int                     `json:"successCount"`
	FailedCount  int                     `json:"failedCount"`
	Rows         []TicketImportRowResult `json:"rows"`
}

type importTicketRow struct {
	LotteryCode      string
	RecommendationID string
	Issue            string
	DrawDate         time.Time
	PurchasedAt      time.Time
	CostAmount       float64
	Notes            string
	EntriesText      string
	RedNumbers       string
	BlueNumbers      string
	Multiple         int
	IsAdditional     bool
	ImageName        string
}

func ImportTickets(ctx context.Context, input ImportTicketsInput) (*TicketImportResult, error) {
	if len(input.Workbook) == 0 {
		return nil, fmt.Errorf("请上传 Excel 文件")
	}

	images, cleanup, err := prepareImportedImages(input.ImagesArchive)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	workbook, err := excelize.OpenReader(bytes.NewReader(input.Workbook))
	if err != nil {
		return nil, fmt.Errorf("Excel 文件无法解析")
	}
	defer workbook.Close()

	sheetName := workbook.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel 中没有可读取的工作表")
	}

	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 行数据失败")
	}
	if len(rows) <= 1 {
		return nil, fmt.Errorf("Excel 中没有可导入的数据")
	}

	headerMap := buildImportHeaderMap(rows[0])
	result := &TicketImportResult{
		Rows: make([]TicketImportRowResult, 0, len(rows)-1),
	}

	for index := 1; index < len(rows); index++ {
		rowNumber := index + 1
		rowInput, skip, parseErr := parseImportTicketRow(rows[index], headerMap)
		if skip {
			continue
		}

		result.TotalCount++
		rowResult := TicketImportRowResult{
			Row:         rowNumber,
			LotteryCode: rowInput.LotteryCode,
			Issue:       rowInput.Issue,
		}

		if parseErr != nil {
			rowResult.Status = "failed"
			rowResult.Message = parseErr.Error()
			result.FailedCount++
			result.Rows = append(result.Rows, rowResult)
			continue
		}

		ticket, importErr := importTicketRowWithImage(ctx, input.UserID, rowInput, images)
		if importErr != nil {
			rowResult.Status = "failed"
			rowResult.Message = importErr.Error()
			result.FailedCount++
			result.Rows = append(result.Rows, rowResult)
			continue
		}

		rowResult.LotteryCode = ticket.LotteryCode
		rowResult.Issue = ticket.Issue
		rowResult.TicketID = ticket.Id.String()
		rowResult.Status = "success"
		result.SuccessCount++
		result.Rows = append(result.Rows, rowResult)
	}

	return result, nil
}

func importTicketRowWithImage(ctx context.Context, userID string, row importTicketRow, images map[string]string) (*TicketDetail, error) {
	recommendation, recommendationID, err := resolveRecommendation(userID, row.LotteryCode, row.RecommendationID)
	if err != nil {
		return nil, err
	}

	code, err := resolveCreateTicketCode(row.LotteryCode, "", recommendation)
	if err != nil {
		return nil, err
	}

	definition, err := GetDefinition(code)
	if err != nil {
		return nil, err
	}

	entries := []ParsedEntry(nil)
	if strings.TrimSpace(row.EntriesText) != "" {
		entries, err = parseImportedEntries(row.EntriesText, code)
		if err != nil {
			return nil, err
		}
	} else {
		entries, err = buildImportedSingleEntry(row, code)
		if err != nil {
			return nil, err
		}
	}
	entries, err = normalizeParsedEntries(definition, entries)
	if err != nil {
		return nil, err
	}

	issue := normalizeIssueByCode(code, row.Issue)
	if issue == "" {
		return nil, fmt.Errorf("期号不能为空")
	}

	imagePath, err := resolveImportedImagePath(row.ImageName, images)
	if err != nil {
		return nil, err
	}

	purchasedAt := row.PurchasedAt
	if purchasedAt.IsZero() {
		purchasedAt = time.Now()
	}

	ticketID := ""
	shouldEvaluate := false
	drawDate := row.DrawDate
	costAmount := row.CostAmount

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := validateDuplicateTicket(tx, userID, code, issue, entries); err != nil {
			return err
		}

		ticket, createErr := createTicketRecord(
			tx,
			userID,
			code,
			recommendationID,
			issue,
			drawDate,
			"import",
			imagePath,
			row.EntriesText,
			purchasedAt,
			costAmount,
			row.Notes,
			entries,
		)
		if createErr != nil {
			if isUniqueConstraintError(createErr) {
				return ErrDuplicateTicket
			}
			return createErr
		}

		ticketID = ticket.Id.String()
		shouldEvaluate = len(entries) > 0
		return nil
	}); err != nil {
		return nil, err
	}

	if shouldDeferSettlement(code, &drawDate) {
		return GetTicketDetail(ticketID, userID)
	}

	ensureIssueDrawSynced(ctx, code, issue)
	if shouldEvaluate {
		if err := EvaluateTicket(ticketID); err != nil {
			return nil, err
		}
	}

	return GetTicketDetail(ticketID, userID)
}

func parseImportTicketRow(row []string, headerMap map[string]int) (importTicketRow, bool, error) {
	item := importTicketRow{
		LotteryCode:      normalizeImportedLotteryCode(readImportCell(row, headerMap, "lotteryCode")),
		RecommendationID: strings.TrimSpace(readImportCell(row, headerMap, "recommendationId")),
		Issue:            strings.TrimSpace(readImportCell(row, headerMap, "issue")),
		Notes:            strings.TrimSpace(readImportCell(row, headerMap, "notes")),
		EntriesText:      strings.TrimSpace(readImportCell(row, headerMap, "entries")),
		RedNumbers:       strings.TrimSpace(readImportCell(row, headerMap, "redNumbers")),
		BlueNumbers:      strings.TrimSpace(readImportCell(row, headerMap, "blueNumbers")),
		ImageName:        strings.TrimSpace(readImportCell(row, headerMap, "imageName")),
	}

	if isEmptyImportRow(item) {
		return item, true, nil
	}
	if item.LotteryCode == "" {
		return item, false, fmt.Errorf("彩票类型不能为空")
	}
	if item.Issue == "" {
		return item, false, fmt.Errorf("期号不能为空")
	}
	if item.EntriesText == "" && (item.RedNumbers == "" || item.BlueNumbers == "") {
		return item, false, fmt.Errorf("号码不能为空，请填写 entries 或红球/蓝球列")
	}

	drawDate, err := parseImportDate(readImportCell(row, headerMap, "drawDate"))
	if err != nil {
		return item, false, fmt.Errorf("开奖日期格式不正确")
	}
	item.DrawDate = drawDate

	purchasedAt, err := parseImportDateTime(readImportCell(row, headerMap, "purchasedAt"))
	if err != nil {
		return item, false, fmt.Errorf("购买时间格式不正确")
	}
	item.PurchasedAt = purchasedAt

	costAmount, err := parseImportFloat(readImportCell(row, headerMap, "costAmount"))
	if err != nil {
		return item, false, fmt.Errorf("金额格式不正确")
	}
	item.CostAmount = costAmount

	multiple, err := parseImportInt(readImportCell(row, headerMap, "multiple"), 1)
	if err != nil {
		return item, false, fmt.Errorf("倍数格式不正确")
	}
	item.Multiple = multiple

	isAdditional, err := parseImportBool(readImportCell(row, headerMap, "isAdditional"))
	if err != nil {
		return item, false, fmt.Errorf("追加格式不正确")
	}
	item.IsAdditional = isAdditional

	return item, false, nil
}

func parseImportedEntries(value string, lotteryCode string) ([]ParsedEntry, error) {
	lines := splitImportedEntries(value)
	if len(lines) == 0 {
		return nil, fmt.Errorf("号码不能为空")
	}

	result := make([]ParsedEntry, 0, len(lines))
	for _, line := range lines {
		entry, err := parseImportedEntryLine(line, lotteryCode)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, nil
}

func buildImportedSingleEntry(row importTicketRow, lotteryCode string) ([]ParsedEntry, error) {
	if row.RedNumbers == "" || row.BlueNumbers == "" {
		return nil, fmt.Errorf("红球和蓝球不能为空")
	}

	return []ParsedEntry{
		{
			Red:          parseCSVNumbers(normalizeImportedNumbers(row.RedNumbers)),
			Blue:         parseCSVNumbers(normalizeImportedNumbers(row.BlueNumbers)),
			Multiple:     max(1, row.Multiple),
			IsAdditional: lotteryCode == "dlt" && row.IsAdditional,
		},
	}, nil
}

func parseImportedEntryLine(line string, lotteryCode string) (ParsedEntry, error) {
	isAdditional := lotteryCode == "dlt" && strings.Contains(line, "追加")
	sourceLine := strings.TrimSpace(strings.ReplaceAll(line, "追加", ""))

	multiple := 1
	if match := regexpMultipleAtEnd.FindStringSubmatch(sourceLine); len(match) == 2 {
		parsed, err := strconv.Atoi(match[1])
		if err == nil && parsed > 0 {
			multiple = parsed
		}
		sourceLine = strings.TrimSpace(regexpMultipleAtEnd.ReplaceAllString(sourceLine, ""))
	}

	parts := strings.Split(sourceLine, "+")
	if len(parts) != 2 {
		return ParsedEntry{}, fmt.Errorf("号码格式不正确，应为 红球+蓝球")
	}

	return ParsedEntry{
		Red:          parseCSVNumbers(normalizeImportedNumbers(parts[0])),
		Blue:         parseCSVNumbers(normalizeImportedNumbers(parts[1])),
		Multiple:     multiple,
		IsAdditional: isAdditional,
	}, nil
}

func splitImportedEntries(value string) []string {
	replacer := strings.NewReplacer("\r\n", "\n", "\r", "\n", "；", "\n", ";", "\n")
	lines := strings.Split(replacer.Replace(value), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func normalizeImportedNumbers(value string) string {
	compact := strings.ReplaceAll(value, " ", "")
	if strings.Contains(compact, ",") {
		return compact
	}
	tokens := numberPattern.FindAllString(compact, -1)
	return strings.Join(tokens, ",")
}

func buildImportHeaderMap(header []string) map[string]int {
	result := make(map[string]int, len(header))
	for index, value := range header {
		key := normalizeImportHeader(value)
		if key != "" {
			result[key] = index
		}
	}
	return result
}

func normalizeImportHeader(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "lotterycode", "lottery", "type", "彩种", "彩票类型", "彩票":
		return "lotteryCode"
	case "recommendationid", "推荐id", "推荐记录id":
		return "recommendationId"
	case "issue", "issueno", "期号":
		return "issue"
	case "drawdate", "开奖日期", "开奖时间":
		return "drawDate"
	case "purchasedat", "购买时间", "录入时间":
		return "purchasedAt"
	case "costamount", "amount", "金额", "花费":
		return "costAmount"
	case "notes", "备注":
		return "notes"
	case "entries", "numbers", "号码", "内容":
		return "entries"
	case "rednumbers", "red", "红球", "前区":
		return "redNumbers"
	case "bluenumbers", "blue", "蓝球", "后区":
		return "blueNumbers"
	case "multiple", "倍数", "注数":
		return "multiple"
	case "isadditional", "additional", "追加", "是否追加":
		return "isAdditional"
	case "imagename", "image", "图片", "图片名", "图片文件名":
		return "imageName"
	default:
		return ""
	}
}

func readImportCell(row []string, headerMap map[string]int, key string) string {
	index, ok := headerMap[key]
	if !ok || index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func isEmptyImportRow(row importTicketRow) bool {
	return row.LotteryCode == "" &&
		row.RecommendationID == "" &&
		row.Issue == "" &&
		row.Notes == "" &&
		row.EntriesText == "" &&
		row.RedNumbers == "" &&
		row.BlueNumbers == "" &&
		row.ImageName == ""
}

func normalizeImportedLotteryCode(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "")
	switch normalized {
	case "ssq", "双色球", "福彩双色球":
		return "ssq"
	case "dlt", "大乐透", "体彩大乐透":
		return "dlt"
	}

	for _, definition := range ListDefinitions() {
		if normalized == strings.ToLower(strings.TrimSpace(definition.Code)) {
			return definition.Code
		}
		name := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(definition.Name, " ", "")))
		if normalized == name {
			return definition.Code
		}
	}
	return ""
}

func parseImportDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", "20060102"} {
		timestamp, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return timestamp, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func parseImportDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006-01-02",
	} {
		timestamp, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return timestamp, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time")
}

func parseImportFloat(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

func parseImportInt(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return fallback, nil
	}
	return parsed, nil
}

func parseImportBool(value string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "0", "false", "no", "n", "否", "不追加":
		return false, nil
	case "1", "true", "yes", "y", "是", "追加":
		return true, nil
	default:
		return false, fmt.Errorf("invalid bool")
	}
}

func prepareImportedImages(archive []byte) (map[string]string, func(), error) {
	cleanup := func() {}
	if len(archive) == 0 {
		return map[string]string{}, cleanup, nil
	}

	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, cleanup, fmt.Errorf("图片压缩包无法解析")
	}

	batchID := time.Now().Format("20060102150405")
	result := make(map[string]string)
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		name := filepath.Base(file.Name)
		if name == "." || name == "" {
			continue
		}

		savedPath := filepath.Join(config.Current.Storage.UploadDir, "imports", batchID, name)
		if err := util.EnsureDir(savedPath); err != nil {
			return nil, cleanup, err
		}

		source, err := file.Open()
		if err != nil {
			return nil, cleanup, err
		}
		target, err := os.Create(savedPath)
		if err != nil {
			source.Close()
			return nil, cleanup, err
		}
		if _, err := io.Copy(target, source); err != nil {
			target.Close()
			source.Close()
			return nil, cleanup, err
		}
		target.Close()
		source.Close()

		result[strings.ToLower(name)] = savedPath
	}
	return result, cleanup, nil
}

func resolveImportedImagePath(imageName string, images map[string]string) (string, error) {
	imageName = strings.TrimSpace(imageName)
	if imageName == "" {
		return "", nil
	}

	path, ok := images[strings.ToLower(filepath.Base(imageName))]
	if !ok {
		return "", fmt.Errorf("图片 %s 未在压缩包中找到", imageName)
	}
	return path, nil
}
