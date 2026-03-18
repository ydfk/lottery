package lottery

import (
	"fmt"
	"strings"
)

type LotteryRecognitionParser interface {
	Code() string
	ParseText(text string) (*RecognitionResult, error)
}

var lotteryRecognitionParsers = map[string]LotteryRecognitionParser{
	"ssq": ssqRecognitionParser{},
	"dlt": dltRecognitionParser{},
}

func ParseLotteryText(code string, text string) (*RecognitionResult, error) {
	parser, err := getLotteryRecognitionParser(code)
	if err != nil {
		return nil, err
	}

	result, err := parser.ParseText(text)
	if err != nil {
		return nil, err
	}
	if result.LotteryCode == "" {
		result.LotteryCode = parser.Code()
	}
	return result, nil
}

func DetectLotteryByText(text string) (*RecognitionResult, error) {
	for _, code := range detectLikelyLotteryCodes(text) {
		result, err := ParseLotteryText(code, text)
		if err == nil && len(result.Entries) > 0 {
			return result, nil
		}
	}

	definitions := ListDefinitions()
	for _, definition := range definitions {
		if !definition.Enabled {
			continue
		}

		result, err := ParseLotteryText(definition.Code, text)
		if err == nil && len(result.Entries) > 0 {
			return result, nil
		}
	}

	return nil, fmt.Errorf("未识别出彩票类型，请补充 OCR 文本或手动选择彩种")
}

func detectLikelyLotteryCodes(text string) []string {
	result := make([]string, 0, 2)
	if strings.Contains(text, "大乐透") {
		result = append(result, "dlt")
	}
	if strings.Contains(text, "双色球") {
		result = append(result, "ssq")
	}
	return result
}

func getLotteryRecognitionParser(code string) (LotteryRecognitionParser, error) {
	parser, ok := lotteryRecognitionParsers[code]
	if !ok {
		return nil, fmt.Errorf("当前暂不支持 %s 的自动识别", code)
	}
	return parser, nil
}
