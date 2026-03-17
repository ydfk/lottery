package lottery

import (
	"fmt"
	"strconv"
	"strings"
)

type ssqRecognitionParser struct{}

func (ssqRecognitionParser) Code() string {
	return "ssq"
}

func (ssqRecognitionParser) ParseText(text string) (*RecognitionResult, error) {
	return ParseSSQText(text)
}

type ParsedEntry struct {
	Red      []int `json:"red"`
	Blue     []int `json:"blue"`
	Multiple int   `json:"multiple"`
}

type RecognitionResult struct {
	LotteryCode string        `json:"lotteryCode"`
	Issue       string        `json:"issue"`
	RawText     string        `json:"rawText"`
	Confidence  float64       `json:"confidence"`
	Entries     []ParsedEntry `json:"entries"`
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
		LotteryCode: "ssq",
		Issue:       parseIssue(text),
		RawText:     text,
		Confidence:  0.6,
		Entries:     normalizeParsedEntriesList(entries),
	}, nil
}

func parseSSQLine(line string) []ParsedEntry {
	multiple := parseEntryMultiple(line)
	line = entryMultiplePattern.ReplaceAllString(line, " ")
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
				Red:      red,
				Blue:     []int{blue},
				Multiple: multiple,
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

func normalizeParsedEntriesList(entries []ParsedEntry) []ParsedEntry {
	result := make([]ParsedEntry, 0, len(entries))
	for _, entry := range entries {
		multiple := entry.Multiple
		if multiple <= 0 {
			multiple = 1
		}
		result = append(result, ParsedEntry{
			Red:      parseCSVNumbers(formatNumbers(entry.Red)),
			Blue:     parseCSVNumbers(formatNumbers(entry.Blue)),
			Multiple: multiple,
		})
	}
	return result
}
