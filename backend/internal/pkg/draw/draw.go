// filepath: d:\1.Development\4.gitea\lottery\backend\internal\pkg\draw\draw.go
package draw

import (
	"encoding/json"
	"fmt"
	"io"
	"lottery-backend/internal/models"
	"lottery-backend/internal/pkg/config"
	"lottery-backend/internal/pkg/database"
	"lottery-backend/internal/pkg/logger"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// LotteryDrawInfo stores information about lottery draws
type LotteryDrawInfo struct {
	Code            string
	Name            string
	CurrentDrawDate time.Time
	CurrentDrawNum  string
	IsDrawToday     bool
	NextDrawDate    time.Time
	NextDrawNum     string
}

// JisuAPIResponse 极速API返回的数据结构
type JisuAPIResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Result struct {
		CaipiaoID        int     `json:"caipiaoid"`        // 彩票ID
		Issueno          string  `json:"issueno"`          // 期号
		Number           string  `json:"number"`           // 开奖号码
		ReferNumber      string  `json:"refernumber"`      // 特殊号码
		OpenDate         string  `json:"opendate"`         // 开奖日期
		OfficialOpenDate string  `json:"officialopendate"` // 官方开奖日期
		Deadline         string  `json:"deadline"`         // 截止日期
		SaleAmount       float64 `json:"saleamount"`       // 销售额
		TotalMoney       string  `json:"totalmoney"`       // 奖池金额
		Prize            []Prize `json:"prize"`            // 奖项信息
	} `json:"result"`
}

// Prize 奖项信息结构体
type Prize struct {
	PrizeName   string  `json:"prizename"`   // 奖项名称
	Require     string  `json:"require"`     // 中奖要求
	Num         int     `json:"num"`         // 中奖注数
	SingleBonus float64 `json:"singlebonus"` // 单注奖金
}

// GetLotteryDrawInfo 获取彩票开奖信息（下一期的信息）
func GetLotteryDrawInfo(lotteryCode string, scheduleCron string) (*LotteryDrawInfo, error) {
	logger.Info("开始获取彩票[%s]开奖信息...", lotteryCode)

	// 1. 获取最新一期的开奖结果
	lotteryType, err := getLotteryTypeByCode(lotteryCode)
	if err != nil {
		return nil, fmt.Errorf("获取彩票类型信息失败: %v", err)
	}

	// 从数据库查询最新一期开奖结果
	latestDraw, err := getLatestDrawResult(lotteryType.Id)
	if err != nil {
		logger.Info("未找到彩票[%s]的历史开奖记录，可能是首次运行", lotteryCode)
	}

	// 根据开奖记录或彩票类型计算下一期信息
	var nextDrawInfo *LotteryDrawInfo
	if latestDraw != nil {
		nextDrawInfo = calculateNextDrawInfo(lotteryType, latestDraw)
	} else {
		// 如果没有历史开奖记录，则根据当前日期和cron表达式计算下一期
		nextDrawInfo = &LotteryDrawInfo{
			Code:            lotteryCode,
			Name:            lotteryType.Name,
			CurrentDrawDate: time.Now(), // 将当前日期作为最近一期日期
		}

		// 根据cron表达式预测下一期开奖日期
		nextDrawDate, err := calculateNextDrawDate(scheduleCron)
		if err != nil {
			logger.Error("计算下一期开奖日期失败: %v", err)
			// 默认为当天晚上8点
			now := time.Now()
			nextDrawDate = time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
		}

		nextDrawInfo.NextDrawDate = nextDrawDate
		nextDrawInfo.NextDrawNum = generateNextDrawNumber(lotteryCode, "", nextDrawDate)
	}

	nextDrawInfo.IsDrawToday = isDrawToday(nextDrawInfo.NextDrawDate)
	return nextDrawInfo, nil
}

// 从数据库获取彩票类型信息
func getLotteryTypeByCode(code string) (*models.LotteryType, error) {
	var lotteryType models.LotteryType
	if err := database.DB.Where("code = ?", code).First(&lotteryType).Error; err != nil {
		return nil, err
	}
	return &lotteryType, nil
}

// 从数据库获取最新一期开奖结果
func getLatestDrawResult(lotteryTypeID uint) (*models.DrawResult, error) {
	var drawResult models.DrawResult
	if err := database.DB.Where("lottery_type_id = ?", lotteryTypeID).
		Order("draw_date DESC").First(&drawResult).Error; err != nil {
		return nil, err
	}
	return &drawResult, nil
}

// 计算下一期开奖信息
func calculateNextDrawInfo(lotteryType *models.LotteryType, latestDraw *models.DrawResult) *LotteryDrawInfo {
	info := &LotteryDrawInfo{
		Code:            lotteryType.Code,
		Name:            lotteryType.Name,
		CurrentDrawDate: latestDraw.DrawDate,
		CurrentDrawNum:  latestDraw.DrawNumber,
	}

	// 根据cron表达式计算下一期开奖日期
	nextDrawDate, err := calculateNextDrawDate(lotteryType.ScheduleCron)
	if err != nil {
		logger.Error("计算下一期开奖日期失败: %v", err)
		// 默认为3天后
		nextDrawDate = latestDraw.DrawDate.AddDate(0, 0, 3)
	}

	info.NextDrawDate = nextDrawDate
	info.NextDrawNum = generateNextDrawNumber(lotteryType.Code, latestDraw.DrawNumber, nextDrawDate)

	return info
}

// 计算下一期开奖日期
func calculateNextDrawDate(cronExpr string) (time.Time, error) {
	// 简化处理：根据当前时间和cron表达式计算下一次执行时间
	// 这里可以使用cron库进行更复杂的计算
	now := time.Now()

	// 根据彩票类型和当前星期几来确定下一期开奖日期
	// 简化处理，不同彩票类型有不同的开奖日期规则
	switch {
	case strings.Contains(cronExpr, "1,3,6"): // 大乐透：周一、三、六
		dayOfWeek := int(now.Weekday())
		daysToAdd := 0

		switch dayOfWeek {
		case 0: // 周日
			daysToAdd = 1 // 下一期是周一
		case 1: // 周一
			daysToAdd = 2 // 下一期是周三
		case 2: // 周二
			daysToAdd = 1 // 下一期是周三
		case 3: // 周三
			daysToAdd = 3 // 下一期是周六
		case 4: // 周四
			daysToAdd = 2 // 下一期是周六
		case 5: // 周五
			daysToAdd = 1 // 下一期是周六
		case 6: // 周六
			daysToAdd = 2 // 下一期是周一
		}

		nextDate := now.AddDate(0, 0, daysToAdd)
		return time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), 20, 0, 0, 0, nextDate.Location()), nil

	case strings.Contains(cronExpr, "2,4,0"): // 双色球：周二、四、日
		dayOfWeek := int(now.Weekday())
		daysToAdd := 0

		switch dayOfWeek {
		case 0: // 周日
			daysToAdd = 2 // 下一期是周二
		case 1: // 周一
			daysToAdd = 1 // 下一期是周二
		case 2: // 周二
			daysToAdd = 2 // 下一期是周四
		case 3: // 周三
			daysToAdd = 1 // 下一期是周四
		case 4: // 周四
			daysToAdd = 3 // 下一期是周日
		case 5: // 周五
			daysToAdd = 2 // 下一期是周日
		case 6: // 周六
			daysToAdd = 1 // 下一期是周日
		}

		nextDate := now.AddDate(0, 0, daysToAdd)
		return time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), 20, 0, 0, 0, nextDate.Location()), nil
	}

	// 默认返回当天晚上8点
	return time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location()), nil
}

// 生成下一期期号
func generateNextDrawNumber(lotteryCode string, currentDrawNumber string, nextDrawDate time.Time) string {
	// 如果没有当前期号，则根据日期生成
	if currentDrawNumber == "" {
		year := nextDrawDate.Format("2006")
		return fmt.Sprintf("%s001", year[2:]) // 如202501
	}

	// 解析当前期号
	numPart := currentDrawNumber
	if len(currentDrawNumber) > 5 {
		numPart = currentDrawNumber[len(currentDrawNumber)-3:] // 取期号的后三位
	}

	yearPart := ""
	if len(currentDrawNumber) >= 5 {
		yearPart = currentDrawNumber[:len(currentDrawNumber)-3] // 取年份部分
	} else {
		yearPart = fmt.Sprintf("%s", nextDrawDate.Format("06")) // 默认使用当前年份
	}

	// 解析数字部分并加1
	num, err := strconv.Atoi(numPart)
	if err != nil {
		// 解析失败，使用默认格式
		return fmt.Sprintf("%s%03d", nextDrawDate.Format("06"), 1)
	}

	// 检查是否需要更新年份部分
	currentYear := nextDrawDate.Format("06")
	if yearPart != currentYear && num >= 150 { // 假设每年至少150期
		// 新的一年，期号重置
		return fmt.Sprintf("%s001", currentYear)
	}

	// 正常情况，期号加1
	return fmt.Sprintf("%s%03d", yearPart, num+1)
}

// 判断是否今天开奖
func isDrawToday(drawDate time.Time) bool {
	today := time.Now().Format("2006-01-02")
	drawDay := drawDate.Format("2006-01-02")
	return today == drawDay
}

// FetchLatestDrawResult 从API获取最新开奖结果
func FetchLatestDrawResult(lotteryType *models.LotteryType) (*models.DrawResult, error) {
	logger.Info("开始从API获取彩票[%s]最新开奖结果...", lotteryType.Name)

	// 构建API请求URL
	apiURL := config.Current.LotteryAPI.BaseURL
	params := url.Values{}
	params.Add("appkey", config.Current.LotteryAPI.AppKey)
	params.Add("caipiaoid", strconv.Itoa(lotteryType.CaipiaoId))

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	logger.Info("请求URL: %s", fullURL)

	// 设置超时时间
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送请求
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加请求头
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回非200状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取API响应失败: %v", err)
	}

	// 解析响应内容
	var apiResp JisuAPIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, fmt.Errorf("解析API响应失败: %v, 原始数据: %s", err, string(body))
	}

	// 检查API返回的状态码
	if apiResp.Status != 0 {
		return nil, fmt.Errorf("API返回错误: %s", apiResp.Msg)
	}

	// 解析开奖日期
	drawDate, err := time.Parse("2006-01-02", apiResp.Result.OpenDate)
	if err != nil {
		return nil, fmt.Errorf("解析开奖日期失败: %v", err)
	}

	// 解析奖池金额
	poolAmount, _ := strconv.ParseFloat(strings.TrimSuffix(apiResp.Result.TotalMoney, ".00"), 64)

	// 将奖项信息转换为JSON
	var prizeJSON []byte
	// 处理prize字段可能为空的情况
	if apiResp.Result.Prize != nil && len(apiResp.Result.Prize) > 0 {
		prizeJSON, err = json.Marshal(apiResp.Result.Prize)
		if err != nil {
			logger.Error("将奖项信息转换为JSON失败: %v", err)
			// 失败不影响整体流程，继续处理
			prizeJSON = []byte("[]") // 使用空数组作为默认值
		}
	} else {
		logger.Warn("彩票[%s]期号[%s]的奖项信息为空", lotteryType.Name, apiResp.Result.Issueno)
		prizeJSON = []byte("[]") // 使用空数组作为默认值
	}

	// 构造开奖结果
	drawResult := &models.DrawResult{
		LotteryTypeID:    lotteryType.Id,
		CaipiaoID:        lotteryType.CaipiaoId,
		DrawNumber:       apiResp.Result.Issueno,
		MainNumbers:      apiResp.Result.Number,
		SpecialNumbers:   apiResp.Result.ReferNumber,
		DrawDate:         drawDate,
		SaleAmount:       apiResp.Result.SaleAmount,
		PoolAmount:       poolAmount,
		OfficialOpenDate: apiResp.Result.OfficialOpenDate,
		Deadline:         apiResp.Result.Deadline,
		PrizeInfo:        models.JSON(prizeJSON),
	}

	// 处理奖项信息
	if apiResp.Result.Prize != nil && len(apiResp.Result.Prize) > 0 {
		processPrizeInfo(drawResult, apiResp.Result.Prize)
	}

	logger.Info("成功获取彩票[%s]期号[%s]开奖结果: %s+%s",
		lotteryType.Name, drawResult.DrawNumber, drawResult.MainNumbers, drawResult.SpecialNumbers)

	return drawResult, nil
}

// processPrizeInfo 处理奖项信息
func processPrizeInfo(drawResult *models.DrawResult, prizes []Prize) {
	for _, prize := range prizes {
		prizeName := prize.PrizeName
		bonus := prize.SingleBonus
		num := prize.Num

		switch {
		case strings.Contains(prizeName, "一等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.FirstPrize = bonus
			drawResult.FirstPrizeNum = num
		case strings.Contains(prizeName, "二等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.SecondPrize = bonus
			drawResult.SecondPrizeNum = num
		case strings.Contains(prizeName, "三等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.ThirdPrize = bonus
			drawResult.ThirdPrizeNum = num
		case strings.Contains(prizeName, "四等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.FourthPrize = bonus
			drawResult.FourthPrizeNum = num
		case strings.Contains(prizeName, "五等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.FifthPrize = bonus
			drawResult.FifthPrizeNum = num
		case strings.Contains(prizeName, "六等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.SixthPrize = bonus
			drawResult.SixthPrizeNum = num
		case strings.Contains(prizeName, "七等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.SeventhPrize = bonus
			drawResult.SeventhPrizeNum = num
		case strings.Contains(prizeName, "八等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.EighthPrize = bonus
			drawResult.EighthPrizeNum = num
		case strings.Contains(prizeName, "九等奖") && !strings.Contains(prizeName, "追加"):
			drawResult.NinthPrize = bonus
			drawResult.NinthPrizeNum = num

		// 处理追加奖项
		case strings.Contains(prizeName, "一等奖") && strings.Contains(prizeName, "追加"):
			drawResult.FirstPrizeAdd = bonus
			drawResult.FirstPrizeAddNum = num
		case strings.Contains(prizeName, "二等奖") && strings.Contains(prizeName, "追加"):
			drawResult.SecondPrizeAdd = bonus
			drawResult.SecondPrizeAddNum = num
		case strings.Contains(prizeName, "三等奖") && strings.Contains(prizeName, "追加"):
			drawResult.ThirdPrizeAdd = bonus
			drawResult.ThirdPrizeAddNum = num
		case strings.Contains(prizeName, "四等奖") && strings.Contains(prizeName, "追加"):
			drawResult.FourthPrizeAdd = bonus
			drawResult.FourthPrizeAddNum = num
		case strings.Contains(prizeName, "五等奖") && strings.Contains(prizeName, "追加"):
			drawResult.FifthPrizeAdd = bonus
			drawResult.FifthPrizeAddNum = num
		case strings.Contains(prizeName, "六等奖") && strings.Contains(prizeName, "追加"):
			drawResult.SixthPrizeAdd = bonus
			drawResult.SixthPrizeAddNum = num
		case strings.Contains(prizeName, "七等奖") && strings.Contains(prizeName, "追加"):
			drawResult.SeventhPrizeAdd = bonus
			drawResult.SeventhPrizeAddNum = num
		}
	}
}

// ProcessDrawResult 处理开奖结果，包括保存结果和分析中奖情况
func ProcessDrawResult(drawResult *models.DrawResult) error {
	// 1. 保存开奖结果到数据库
	if err := saveDrawResult(drawResult); err != nil {
		return err
	}

	// 2. 查找该期号对应的所有推荐记录
	var recommendations []models.Recommendation
	if err := database.DB.Where("lottery_type_id = ? AND draw_number = ?",
		drawResult.LotteryTypeID, drawResult.DrawNumber).Find(&recommendations).Error; err != nil {
		return fmt.Errorf("查询推荐记录失败: %v", err)
	}

	// 3. 分析每条推荐记录的中奖情况
	logger.Info("开始分析彩票类型[ID:%d]期号[%s]的推荐记录中奖情况，共%d条记录",
		drawResult.LotteryTypeID, drawResult.DrawNumber, len(recommendations))

	for i, rec := range recommendations {
		// 如果记录已经有中奖分析结果且状态不是"未知"，则跳过
		if rec.DrawResult != "" && rec.WinStatus != "" && rec.WinStatus != "未知" {
			logger.Info("推荐记录[ID:%d]已有中奖分析结果：%s，跳过分析", rec.Id, rec.WinStatus)
			continue
		}

		winStatus, winAmount := analyzeWinResult(&rec, drawResult)

		// 更新推荐记录的中奖情况
		rec.DrawResult = fmt.Sprintf("%s+%s", drawResult.MainNumbers, drawResult.SpecialNumbers)
		rec.WinStatus = winStatus
		rec.WinAmount = winAmount

		if err := database.DB.Save(&rec).Error; err != nil {
			logger.Error("更新推荐记录[ID:%d]中奖情况失败: %v", rec.Id, err)
			continue
		}

		logger.Info("推荐记录[ID:%d](%d/%d)中奖分析完成: 状态=%s, 金额=%.2f",
			rec.Id, i+1, len(recommendations), winStatus, winAmount)
	}

	logger.Info("彩票类型[ID:%d]期号[%s]的所有推荐记录中奖分析完成，共处理%d条记录",
		drawResult.LotteryTypeID, drawResult.DrawNumber, len(recommendations))

	return nil
}

// 保存开奖结果到数据库
func saveDrawResult(drawResult *models.DrawResult) error {
	// 先查询是否已存在
	var count int64
	database.DB.Model(&models.DrawResult{}).Where(
		"lottery_type_id = ? AND draw_number = ?",
		drawResult.LotteryTypeID,
		drawResult.DrawNumber,
	).Count(&count)

	if count > 0 {
		logger.Info("彩票类型[ID:%d]期号[%s]的开奖结果已存在，更新数据",
			drawResult.LotteryTypeID, drawResult.DrawNumber)
		return database.DB.Model(&models.DrawResult{}).Where(
			"lottery_type_id = ? AND draw_number = ?",
			drawResult.LotteryTypeID,
			drawResult.DrawNumber,
		).Updates(map[string]interface{}{
			"main_numbers":          drawResult.MainNumbers,
			"special_numbers":       drawResult.SpecialNumbers,
			"draw_date":             drawResult.DrawDate,
			"sale_amount":           drawResult.SaleAmount,
			"pool_amount":           drawResult.PoolAmount,
			"prize_info":            drawResult.PrizeInfo,
			"official_open_date":    drawResult.OfficialOpenDate,
			"deadline":              drawResult.Deadline,
			"first_prize":           drawResult.FirstPrize,
			"first_prize_num":       drawResult.FirstPrizeNum,
			"second_prize":          drawResult.SecondPrize,
			"second_prize_num":      drawResult.SecondPrizeNum,
			"third_prize":           drawResult.ThirdPrize,
			"third_prize_num":       drawResult.ThirdPrizeNum,
			"fourth_prize":          drawResult.FourthPrize,
			"fourth_prize_num":      drawResult.FourthPrizeNum,
			"fifth_prize":           drawResult.FifthPrize,
			"fifth_prize_num":       drawResult.FifthPrizeNum,
			"sixth_prize":           drawResult.SixthPrize,
			"sixth_prize_num":       drawResult.SixthPrizeNum,
			"seventh_prize":         drawResult.SeventhPrize,
			"seventh_prize_num":     drawResult.SeventhPrizeNum,
			"eighth_prize":          drawResult.EighthPrize,
			"eighth_prize_num":      drawResult.EighthPrizeNum,
			"ninth_prize":           drawResult.NinthPrize,
			"ninth_prize_num":       drawResult.NinthPrizeNum,
			"first_prize_add":       drawResult.FirstPrizeAdd,
			"first_prize_add_num":   drawResult.FirstPrizeAddNum,
			"second_prize_add":      drawResult.SecondPrizeAdd,
			"second_prize_add_num":  drawResult.SecondPrizeAddNum,
			"third_prize_add":       drawResult.ThirdPrizeAdd,
			"third_prize_add_num":   drawResult.ThirdPrizeAddNum,
			"fourth_prize_add":      drawResult.FourthPrizeAdd,
			"fourth_prize_add_num":  drawResult.FourthPrizeAddNum,
			"fifth_prize_add":       drawResult.FifthPrizeAdd,
			"fifth_prize_add_num":   drawResult.FifthPrizeAddNum,
			"sixth_prize_add":       drawResult.SixthPrizeAdd,
			"sixth_prize_add_num":   drawResult.SixthPrizeAddNum,
			"seventh_prize_add":     drawResult.SeventhPrizeAdd,
			"seventh_prize_add_num": drawResult.SeventhPrizeAddNum,
		}).Error
	}

	// 如果不存在则创建新记录
	logger.Info("保存新的开奖结果记录：类型ID[%d]期号[%s]", drawResult.LotteryTypeID, drawResult.DrawNumber)
	return database.DB.Create(drawResult).Error
}

// 分析中奖情况
func analyzeWinResult(recommendation *models.Recommendation, drawResult *models.DrawResult) (string, float64) {
	// 获取彩票类型信息
	var lotteryType models.LotteryType
	if err := database.DB.First(&lotteryType, drawResult.LotteryTypeID).Error; err != nil {
		logger.Error("获取彩票类型信息失败: %v", err)
		return "未知", 0
	}

	switch lotteryType.Code {
	case "fc_ssq":
		return analyzeSSQWin(recommendation.Numbers, drawResult.MainNumbers, drawResult.SpecialNumbers)
	case "tc_dlt":
		return analyzeDLTWin(recommendation.Numbers, drawResult.MainNumbers, drawResult.SpecialNumbers)
	default:
		return "未知", 0
	}
}

// 分析双色球中奖情况
func analyzeSSQWin(recommendNumbers, mainNumbers, specialNumbers string) (string, float64) {
	// 解析推荐号码
	parts := strings.Split(recommendNumbers, "+")
	if len(parts) != 2 {
		return "格式错误", 0
	}
	recRedBalls := strings.Split(parts[0], ",")
	recBlueBall := parts[1]

	// 解析开奖号码
	drawRedBalls := strings.Split(mainNumbers, " ")
	drawBlueBall := specialNumbers

	// 计算红蓝球匹配数量
	redMatches := countMatches(recRedBalls, drawRedBalls)
	blueMatch := recBlueBall == drawBlueBall

	// 根据匹配情况确定奖项
	winLevel := "未中奖"
	winAmount := 0.0

	if redMatches == 6 && blueMatch {
		winLevel = "一等奖"
		winAmount = 5000000 // 估算值，实际奖金根据销量和奖池决定
	} else if redMatches == 6 {
		winLevel = "二等奖"
		winAmount = 100000
	} else if redMatches == 5 && blueMatch {
		winLevel = "三等奖"
		winAmount = 3000
	} else if redMatches == 5 || (redMatches == 4 && blueMatch) {
		winLevel = "四等奖"
		winAmount = 200
	} else if redMatches == 4 || (redMatches == 3 && blueMatch) {
		winLevel = "五等奖"
		winAmount = 10
	} else if blueMatch {
		winLevel = "六等奖"
		winAmount = 5
	}

	return winLevel, winAmount
}

// 分析大乐透中奖情况
func analyzeDLTWin(recommendNumbers, mainNumbers, specialNumbers string) (string, float64) {
	// 解析推荐号码
	parts := strings.Split(recommendNumbers, "+")
	if len(parts) != 2 {
		return "格式错误", 0
	}
	recFrontBalls := strings.Split(parts[0], ",")
	recBackBalls := strings.Split(parts[1], ",")

	// 解析开奖号码
	drawFrontBalls := strings.Split(mainNumbers, " ")
	drawBackBalls := strings.Split(specialNumbers, " ")

	// 计算前后区匹配数量
	frontMatches := countMatches(recFrontBalls, drawFrontBalls)
	backMatches := countMatches(recBackBalls, drawBackBalls)

	// 根据匹配情况确定奖项
	winLevel := "未中奖"
	winAmount := 0.0

	if frontMatches == 5 && backMatches == 2 {
		winLevel = "一等奖"
		winAmount = 10000000 // 估算值，实际奖金根据销量和奖池决定
	} else if frontMatches == 5 && backMatches == 1 {
		winLevel = "二等奖"
		winAmount = 200000
	} else if frontMatches == 5 && backMatches == 0 {
		winLevel = "三等奖"
		winAmount = 10000
	} else if frontMatches == 4 && backMatches == 2 {
		winLevel = "四等奖"
		winAmount = 3000
	} else if frontMatches == 4 && backMatches == 1 {
		winLevel = "五等奖"
		winAmount = 300
	} else if frontMatches == 3 && backMatches == 2 {
		winLevel = "六等奖"
		winAmount = 200
	} else if frontMatches == 4 && backMatches == 0 {
		winLevel = "七等奖"
		winAmount = 100
	} else if (frontMatches == 3 && backMatches == 1) || (frontMatches == 2 && backMatches == 2) {
		winLevel = "八等奖"
		winAmount = 15
	} else if (frontMatches == 3 && backMatches == 0) || (frontMatches == 1 && backMatches == 2) ||
		(frontMatches == 2 && backMatches == 1) || (frontMatches == 0 && backMatches == 2) {
		winLevel = "九等奖"
		winAmount = 5
	}

	return winLevel, winAmount
}

// 计算两个字符串数组中相同元素的数量
func countMatches(arr1, arr2 []string) int {
	matches := 0
	for _, val1 := range arr1 {
		for _, val2 := range arr2 {
			if val1 == val2 {
				matches++
				break
			}
		}
	}
	return matches
}

// FetchAllActiveLotteryDrawResults 爬取所有活跃彩票类型的开奖结果
func FetchAllActiveLotteryDrawResults() error {
	logger.Info("开始爬取所有活跃彩票类型的开奖结果...")

	// 查询所有活跃的彩票类型
	var lotteryTypes []models.LotteryType
	if err := database.DB.Where("is_active = ?", true).Find(&lotteryTypes).Error; err != nil {
		return fmt.Errorf("查询活跃彩票类型失败: %v", err)
	}

	logger.Info("找到%d个活跃彩票类型", len(lotteryTypes))

	// 遍历每个彩票类型，获取最新开奖结果
	for _, lt := range lotteryTypes {
		logger.Info("开始处理彩票类型[%s]...", lt.Name)

		// 只处理配置了彩票ID的类型
		if lt.CaipiaoId <= 0 {
			logger.Info("彩票类型[%s]未配置彩票ID，跳过处理", lt.Name)
			continue
		}

		// 获取最新开奖结果
		drawResult, err := FetchLatestDrawResult(&lt)
		if err != nil {
			logger.Error("获取彩票[%s]开奖结果失败: %v", lt.Name, err)
			continue
		}

		// 处理开奖结果(保存结果并分析中奖情况)
		if err := ProcessDrawResult(drawResult); err != nil {
			logger.Error("处理彩票[%s]开奖结果失败: %v", lt.Name, err)
			continue
		}

		logger.Info("彩票类型[%s]开奖结果处理完成", lt.Name)
	}

	logger.Info("所有彩票类型开奖结果爬取完成")
	return nil
}
