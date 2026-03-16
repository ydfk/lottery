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

	"go-fiber-starter/pkg/config"
)

var (
	jsonFencePattern = regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
	issuePattern     = regexp.MustCompile(`20\d{5}`)
	numberPattern    = regexp.MustCompile(`\d{1,2}`)
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
	return replacer.Replace(text)
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

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
