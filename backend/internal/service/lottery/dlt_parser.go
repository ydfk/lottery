package lottery

import (
	"fmt"
	"regexp"
	"strings"
)

type dltRecognitionParser struct{}

var dltPackedLinePattern = regexp.MustCompile(`(\d{10})\s*\+\s*(\d{4})`)

func (dltRecognitionParser) Code() string {
	return "dlt"
}

func (dltRecognitionParser) ParseText(text string) (*RecognitionResult, error) {
	return ParseDLTText(text)
}

func ParseDLTText(text string) (*RecognitionResult, error) {
	normalized := normalizeText(text)
	lines := strings.Split(normalized, "\n")
	entries := make([]ParsedEntry, 0)
	multiple := parseTicketMultiple(text)
	isAdditional := strings.Contains(text, "追加")

	for _, line := range lines {
		entries = append(entries, parseDLTLine(line, multiple, isAdditional)...)
	}
	if len(entries) == 0 {
		entries = append(entries, parseDLTLine(normalized, multiple, isAdditional)...)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("未识别到有效的大乐透号码，请补充 OCR 文本后重试")
	}

	return &RecognitionResult{
		LotteryCode: "dlt",
		Issue:       parseIssue(text),
		DrawDate:    parseRecognizedDrawDate(text),
		CostAmount:  parseRecognizedCost(text, entries),
		RawText:     text,
		Confidence:  0.6,
		Entries:     normalizeParsedEntriesList(entries),
	}, nil
}

func parseDLTLine(line string, fallbackMultiple int, isAdditional bool) []ParsedEntry {
	if !strings.Contains(line, "+") {
		return nil
	}

	if entries := parsePackedDLTEntries(line, fallbackMultiple, isAdditional); len(entries) > 0 {
		return entries
	}

	multiple := fallbackMultiple
	lineMultiple := parseEntryMultiple(line)
	if lineMultiple > 1 {
		multiple = lineMultiple
	}

	cleaned := entryMultiplePattern.ReplaceAllString(line, " ")
	tokens := numberPattern.FindAllString(cleaned, -1)
	if len(tokens) < 7 {
		return nil
	}

	entries := make([]ParsedEntry, 0)
	for index := 0; index+7 <= len(tokens); index++ {
		red := parseTokenSlice(tokens[index : index+5])
		blue := parseTokenSlice(tokens[index+5 : index+7])
		if isValidDLTEntry(red, blue) {
			entries = append(entries, ParsedEntry{
				Red:          red,
				Blue:         blue,
				Multiple:     multiple,
				IsAdditional: isAdditional,
			})
			index += 6
		}
	}
	return entries
}

func parsePackedDLTEntries(line string, fallbackMultiple int, isAdditional bool) []ParsedEntry {
	multiple := fallbackMultiple
	lineMultiple := parseEntryMultiple(line)
	if lineMultiple > 1 {
		multiple = lineMultiple
	}

	matches := dltPackedLinePattern.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return nil
	}

	entries := make([]ParsedEntry, 0, len(matches))
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		red := splitPackedNumbers(match[1], 5)
		blue := splitPackedNumbers(match[2], 2)
		if !isValidDLTEntry(red, blue) {
			continue
		}

		entries = append(entries, ParsedEntry{
			Red:          red,
			Blue:         blue,
			Multiple:     multiple,
			IsAdditional: isAdditional,
		})
	}
	return entries
}

func splitPackedNumbers(value string, count int) []int {
	if len(value) != count*2 {
		return nil
	}

	result := make([]int, 0, count)
	for index := 0; index < len(value); index += 2 {
		result = append(result, parseInt(value[index:index+2]))
	}
	return result
}

func parseTokenSlice(tokens []string) []int {
	result := make([]int, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, parseInt(token))
	}
	return result
}

func isValidDLTEntry(red []int, blue []int) bool {
	if len(red) != 5 || len(blue) != 2 {
		return false
	}
	if containsDuplicate(red) || containsDuplicate(blue) {
		return false
	}
	for _, number := range red {
		if number < 1 || number > 35 {
			return false
		}
	}
	for _, number := range blue {
		if number < 1 || number > 12 {
			return false
		}
	}
	return true
}
