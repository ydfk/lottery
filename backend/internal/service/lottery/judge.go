package lottery

import (
	"fmt"

	model "go-fiber-starter/internal/model/lottery"
)

type PrizeResult struct {
	IsWinning    bool    `json:"isWinning"`
	PrizeName    string  `json:"prizeName"`
	PrizeAmount  float64 `json:"prizeAmount"`
	MatchSummary string  `json:"matchSummary"`
}

func JudgeNumbers(code string, redNumbers string, blueNumbers string, isAdditional bool, draw model.DrawResult, prizeMap map[string]float64) PrizeResult {
	switch code {
	case "dlt":
		return judgeDLT(redNumbers, blueNumbers, isAdditional, draw, prizeMap)
	default:
		return judgeSSQ(redNumbers, blueNumbers, draw, prizeMap)
	}
}

func judgeSSQ(redNumbers string, blueNumbers string, draw model.DrawResult, prizeMap map[string]float64) PrizeResult {
	redHit := countHit(parseCSVNumbers(redNumbers), parseCSVNumbers(draw.RedNumbers))
	blueHit := countHit(parseCSVNumbers(blueNumbers), parseCSVNumbers(draw.BlueNumbers))
	prizeName := resolveSSQPrizeName(redHit, blueHit > 0)

	result := PrizeResult{
		MatchSummary: fmt.Sprintf("%d红%d蓝", redHit, blueHit),
	}
	if prizeName == "" {
		return result
	}

	result.IsWinning = true
	result.PrizeName = prizeName
	result.PrizeAmount = resolvePrizeAmount("ssq", prizeName, prizeMap)
	return result
}

func judgeDLT(redNumbers string, blueNumbers string, isAdditional bool, draw model.DrawResult, prizeMap map[string]float64) PrizeResult {
	redHit := countHit(parseCSVNumbers(redNumbers), parseCSVNumbers(draw.RedNumbers))
	blueHit := countHit(parseCSVNumbers(blueNumbers), parseCSVNumbers(draw.BlueNumbers))
	prizeName := resolveDLTPrizeName(redHit, blueHit)

	result := PrizeResult{
		MatchSummary: fmt.Sprintf("%d前%d后", redHit, blueHit),
	}
	if prizeName == "" {
		return result
	}

	result.IsWinning = true
	result.PrizeName = prizeName
	result.PrizeAmount = resolveDLTPrizeAmount(prizeName, isAdditional, prizeMap)
	return result
}

func countHit(selected []int, target []int) int {
	targetSet := make(map[int]struct{}, len(target))
	for _, number := range target {
		targetSet[number] = struct{}{}
	}

	hitCount := 0
	for _, number := range selected {
		if _, ok := targetSet[number]; ok {
			hitCount++
		}
	}
	return hitCount
}

func resolveSSQPrizeName(redHit int, blueHit bool) string {
	switch {
	case redHit == 6 && blueHit:
		return "一等奖"
	case redHit == 6:
		return "二等奖"
	case redHit == 5 && blueHit:
		return "三等奖"
	case redHit == 5 || (redHit == 4 && blueHit):
		return "四等奖"
	case redHit == 4 || (redHit == 3 && blueHit):
		return "五等奖"
	case blueHit:
		return "六等奖"
	default:
		return ""
	}
}

func resolveDLTPrizeName(redHit int, blueHit int) string {
	switch {
	case redHit == 5 && blueHit == 2:
		return "一等奖"
	case redHit == 5 && blueHit == 1:
		return "二等奖"
	case redHit == 5 && blueHit == 0:
		return "三等奖"
	case redHit == 4 && blueHit == 2:
		return "三等奖"
	case redHit == 4 && blueHit == 1:
		return "四等奖"
	case redHit == 3 && blueHit == 2:
		return "五等奖"
	case redHit == 4 && blueHit == 0:
		return "五等奖"
	case (redHit == 3 && blueHit == 1) || (redHit == 2 && blueHit == 2):
		return "六等奖"
	case (redHit == 3 && blueHit == 0) || (redHit == 2 && blueHit == 1) || (redHit == 1 && blueHit == 2) || (redHit == 0 && blueHit == 2):
		return "七等奖"
	default:
		return ""
	}
}

func resolveDLTPrizeAmount(prizeName string, isAdditional bool, prizeMap map[string]float64) float64 {
	amount := resolvePrizeAmount("dlt", prizeName, prizeMap)
	if !isAdditional {
		return amount
	}
	if prizeName == "一等奖" || prizeName == "二等奖" {
		return amount * 1.8
	}
	return amount
}

func resolvePrizeAmount(code string, prizeName string, prizeMap map[string]float64) float64 {
	if amount, ok := prizeMap[prizeName]; ok && amount > 0 {
		return amount
	}

	switch code {
	case "dlt":
		switch prizeName {
		case "三等奖":
			return 5000
		case "四等奖":
			return 300
		case "五等奖":
			return 150
		case "六等奖":
			return 15
		case "七等奖":
			return 5
		default:
			return 0
		}
	default:
		switch prizeName {
		case "三等奖":
			return 3000
		case "四等奖":
			return 200
		case "五等奖":
			return 10
		case "六等奖":
			return 5
		default:
			return 0
		}
	}
}
