package lottery

import (
	"fmt"
	"strconv"
	"strings"
)

type ParsedEntry struct {
	Red  []int `json:"red"`
	Blue []int `json:"blue"`
}

type RecognitionResult struct {
	Issue      string        `json:"issue"`
	RawText    string        `json:"rawText"`
	Confidence float64       `json:"confidence"`
	Entries    []ParsedEntry `json:"entries"`
}

func ParseSSQText(text string) (*RecognitionResult, error) {
	normalized := normalizeText(text)
	lines := strings.Split(normalized, "\n")
	entries := make([]ParsedEntry, 0)

	for _, line := range lines {
		entries = append(entries, parseSSQLine(line)...)
	}
	if len(entries) == 0 {
		entries = append(entries, parseSSQLine(normalized)...)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("未识别到有效的双色球号码，请补充 OCR 文本后重试")
	}

	return &RecognitionResult{
		Issue:      parseIssue(text),
		RawText:    text,
		Confidence: 0.6,
		Entries:    uniqueEntries(entries),
	}, nil
}

func parseSSQLine(line string) []ParsedEntry {
	tokens := numberPattern.FindAllString(line, -1)
	if len(tokens) < 7 {
		return nil
	}

	numbers := make([]int, 0, len(tokens))
	for _, token := range tokens {
		number, err := strconv.Atoi(token)
		if err != nil {
			continue
		}
		numbers = append(numbers, number)
	}

	entries := make([]ParsedEntry, 0)
	for index := 0; index+7 <= len(numbers); {
		red := append([]int(nil), numbers[index:index+6]...)
		blue := numbers[index+6]
		if isValidSSQEntry(red, blue) {
			entries = append(entries, ParsedEntry{
				Red:  red,
				Blue: []int{blue},
			})
			index += 7
			continue
		}
		index++
	}
	return entries
}

func isValidSSQEntry(red []int, blue int) bool {
	if len(red) != 6 || containsDuplicate(red) {
		return false
	}
	for _, number := range red {
		if number < 1 || number > 33 {
			return false
		}
	}
	return blue >= 1 && blue <= 16
}

func uniqueEntries(entries []ParsedEntry) []ParsedEntry {
	seen := make(map[string]struct{}, len(entries))
	result := make([]ParsedEntry, 0, len(entries))
	for _, entry := range entries {
		key := formatNumbers(entry.Red) + "|" + formatNumbers(entry.Blue)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, ParsedEntry{
			Red:  parseCSVNumbers(formatNumbers(entry.Red)),
			Blue: parseCSVNumbers(formatNumbers(entry.Blue)),
		})
	}
	return result
}
