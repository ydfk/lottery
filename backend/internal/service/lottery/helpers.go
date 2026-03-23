package lottery

import (
	"encoding/json"
	"fmt"
	"math"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"go-fiber-starter/pkg/config"
)

var (
	jsonFencePattern      = regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
	issuePattern          = regexp.MustCompile(`20\d{5}`)
	issueLabelPattern     = regexp.MustCompile(`第\s*(\d{5,7})\s*期`)
	datePattern           = regexp.MustCompile(`(20\d{2})[年./-](\d{1,2})[月./-](\d{1,2})`)
	shortDatePattern      = regexp.MustCompile(`\b(\d{2})[./-](\d{1,2})[./-](\d{1,2})\b`)
	compactDatePattern    = regexp.MustCompile(`\b(20\d{2})(\d{2})(\d{2})\b`)
	costPattern           = regexp.MustCompile(`(?:金额|合计|共计|投注金额|投注额|总金额|实付)\s*[:：]?\s*(\d+(?:\.\d+)?)`)
	numberPattern         = regexp.MustCompile(`\d{1,2}`)
	entryMultiplePattern  = regexp.MustCompile(`[（(]\s*(\d+)\s*[)）]`)
	ticketMultiplePattern = regexp.MustCompile(`(?:追加投注|投注)?\s*(\d+)\s*倍`)
	entryMarkerPattern    = regexp.MustCompile(`([①②③④⑤⑥⑦⑧⑨⑩])`)
)

func mustJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func stripJSONFence(content string) string {
	matches := jsonFencePattern.FindStringSubmatch(content)
	if len(matches) == 2 {
		return matches[1]
	}
	return strings.TrimSpace(content)
}

func parseIssue(text string) string {
	if matches := issueLabelPattern.FindStringSubmatch(text); len(matches) == 2 {
		return matches[1]
	}
	return issuePattern.FindString(text)
}

func normalizeText(text string) string {
	replacer := strings.NewReplacer(
		"，", " ",
		"、", " ",
		"；", "\n",
		";", "\n",
		"|", "\n",
		"｜", "\n",
		",", " ",
		"红", " ",
		"蓝", " ",
		":", " ",
		"：", " ",
	)
	return entryMarkerPattern.ReplaceAllString(replacer.Replace(text), "\n$1")
}

func parseEntryMultiple(text string) int {
	matches := entryMultiplePattern.FindStringSubmatch(text)
	if len(matches) != 2 {
		return 1
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil || value <= 0 {
		return 1
	}
	return value
}

func parseTicketMultiple(text string) int {
	matches := ticketMultiplePattern.FindStringSubmatch(text)
	if len(matches) != 2 {
		return 1
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil || value <= 0 {
		return 1
	}
	return value
}

func parseRecognizedDrawDate(text string) string {
	if matches := datePattern.FindStringSubmatch(text); len(matches) == 4 {
		return formatDateString(matches[1], matches[2], matches[3])
	}
	if matches := shortDatePattern.FindStringSubmatch(text); len(matches) == 4 {
		return formatDateString(normalizeShortYear(matches[1]), matches[2], matches[3])
	}
	if matches := compactDatePattern.FindStringSubmatch(text); len(matches) == 4 {
		return formatDateString(matches[1], matches[2], matches[3])
	}
	return ""
}

func normalizeShortYear(value string) string {
	year, err := strconv.Atoi(value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("20%02d", year)
}

func formatDateString(year string, month string, day string) string {
	monthValue, err := strconv.Atoi(month)
	if err != nil || monthValue < 1 || monthValue > 12 {
		return ""
	}
	dayValue, err := strconv.Atoi(day)
	if err != nil || dayValue < 1 || dayValue > 31 {
		return ""
	}
	return fmt.Sprintf("%s-%02d-%02d", year, monthValue, dayValue)
}

func parseRecognizedCost(text string, entries []ParsedEntry) float64 {
	if matches := costPattern.FindStringSubmatch(text); len(matches) == 2 {
		return parseFloat(matches[1])
	}
	return calculateEntriesCost(entries)
}

func calculateEntriesCost(entries []ParsedEntry) float64 {
	total := 0.0
	for _, entry := range entries {
		multiple := resolveEntryMultiple(entry)
		perBetCost := 2
		if entry.IsAdditional {
			perBetCost++
		}
		total += float64(multiple * perBetCost)
	}
	return total
}

func parseFloat(value any) float64 {
	switch actual := value.(type) {
	case float64:
		return actual
	case float32:
		return float64(actual)
	case int:
		return float64(actual)
	case int64:
		return float64(actual)
	case json.Number:
		number, _ := actual.Float64()
		return number
	case string:
		clean := strings.ReplaceAll(actual, ",", "")
		number, _ := strconv.ParseFloat(clean, 64)
		return number
	default:
		return 0
	}
}

func parseFloatValues(values ...any) float64 {
	for _, value := range values {
		if number := parseFloat(value); number > 0 {
			return number
		}
	}
	return 0
}

func parseInt(value any) int {
	return int(math.Round(parseFloat(value)))
}

func extractString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch actual := value.(type) {
		case string:
			if actual != "" {
				return actual
			}
		case json.Number:
			return actual.String()
		default:
			return fmt.Sprintf("%v", actual)
		}
	}
	return ""
}

func extractItems(raw json.RawMessage) ([]map[string]any, error) {
	var directItems []map[string]any
	if err := json.Unmarshal(raw, &directItems); err == nil && len(directItems) > 0 {
		return directItems, nil
	}

	var wrapper map[string]any
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}

	for _, key := range []string{"list", "data", "result"} {
		value, ok := wrapper[key]
		if !ok {
			continue
		}
		items, ok := value.([]any)
		if !ok {
			continue
		}
		result := make([]map[string]any, 0, len(items))
		for _, item := range items {
			record, ok := item.(map[string]any)
			if ok {
				result = append(result, record)
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}

	return nil, fmt.Errorf("开奖数据格式无法识别")
}

func extractSingleItem(raw json.RawMessage) (map[string]any, error) {
	item := map[string]any{}
	if err := json.Unmarshal(raw, &item); err == nil && len(item) > 0 {
		return item, nil
	}

	var wrapper map[string]any
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}

	for _, key := range []string{"result", "data"} {
		value, ok := wrapper[key]
		if !ok {
			continue
		}
		record, ok := value.(map[string]any)
		if ok {
			return record, nil
		}
	}

	return nil, fmt.Errorf("当前开奖数据格式无法识别")
}

func parseCSVNumbers(value string) []int {
	items := strings.Split(value, ",")
	result := make([]int, 0, len(items))
	for _, item := range items {
		number, err := strconv.Atoi(strings.TrimSpace(item))
		if err == nil {
			result = append(result, number)
		}
	}
	return result
}

func formatNumbers(numbers []int) string {
	cloned := append([]int(nil), numbers...)
	slices.Sort(cloned)
	parts := make([]string, 0, len(cloned))
	for _, number := range cloned {
		parts = append(parts, fmt.Sprintf("%02d", number))
	}
	return strings.Join(parts, ",")
}

func containsDuplicate(numbers []int) bool {
	seen := make(map[int]struct{}, len(numbers))
	for _, number := range numbers {
		if _, ok := seen[number]; ok {
			return true
		}
		seen[number] = struct{}{}
	}
	return false
}

func buildPublicImageURL(imagePath string) string {
	if imagePath == "" {
		return ""
	}
	relativePath, err := filepath.Rel(config.Current.Storage.UploadDir, imagePath)
	if err != nil {
		return ""
	}
	return path.Join("/uploads", filepath.ToSlash(relativePath))
}

func normalizeIssueByCode(code string, issue string) string {
	issue = strings.TrimSpace(issue)
	if issue == "" {
		return ""
	}

	if code == "dlt" && len(issue) == 5 && isDigits(issue) {
		return "20" + issue
	}
	return issue
}

func issueAliases(code string, issue string) []string {
	canonical := normalizeIssueByCode(code, issue)
	aliases := make([]string, 0, 2)
	for _, item := range []string{canonical, strings.TrimSpace(issue)} {
		if item == "" {
			continue
		}
		exists := false
		for _, current := range aliases {
			if current == item {
				exists = true
				break
			}
		}
		if !exists {
			aliases = append(aliases, item)
		}
	}
	if code == "dlt" && len(canonical) == 7 && strings.HasPrefix(canonical, "20") {
		shortIssue := canonical[2:]
		exists := false
		for _, current := range aliases {
			if current == shortIssue {
				exists = true
				break
			}
		}
		if !exists {
			aliases = append(aliases, shortIssue)
		}
	}
	return aliases
}

func isDigits(value string) bool {
	for _, item := range value {
		if item < '0' || item > '9' {
			return false
		}
	}
	return value != ""
}

func normalizeDateOnly(value time.Time) time.Time {
	localValue := value.In(time.Local)
	year, month, day := localValue.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func shouldDeferSettlement(code string, drawDate *time.Time) bool {
	if drawDate == nil || drawDate.IsZero() {
		return false
	}

	now := time.Now()
	targetDate := normalizeDateOnly(*drawDate)
	currentDate := normalizeDateOnly(now)
	if targetDate.After(currentDate) {
		return true
	}
	if !targetDate.Equal(currentDate) {
		return false
	}

	definition, err := GetDefinition(code)
	if err != nil || definition.DrawSchedule.Time == "" {
		return false
	}

	scheduledClock, err := time.ParseInLocation("15:04", definition.DrawSchedule.Time, time.Local)
	if err != nil {
		return false
	}

	return now.Before(time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		scheduledClock.Hour(),
		scheduledClock.Minute(),
		0,
		0,
		time.Local,
	))
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
