// filepath: d:\1.Development\4.gitea\lottery\backend\internal\pkg\draw\draw.go
package draw

import (
	"encoding/json"
	"fmt"
	"io"
	"lottery-backend/internal/pkg/logger"
	"net/http"
	"strings"
	"time"
)

// WeekDay mapping between time.Weekday and cron expression days
var WeekDay = map[time.Weekday]string{
	time.Sunday:    "0",
	time.Monday:    "1",
	time.Tuesday:   "2",
	time.Wednesday: "3",
	time.Thursday:  "4",
	time.Friday:    "5",
	time.Saturday:  "6",
}

// LotteryDrawInfo stores information about lottery draws
type LotteryDrawInfo struct {
	Code            string
	Name            string
	CurrentDrawDate time.Time
	CurrentDrawNum  string
	IsDrawToday     bool
}

// 通用彩票API返回数据结构
type LotteryAPIResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        DrawDate   string `json:"drawDate"`   // 开奖日期
        DrawNumber string `json:"drawNumber"` // 期号
    } `json:"data"`
}

// 体彩大乐透API返回数据结构
type SportLotteryResponse struct {
    DataFrom      string `json:"dataFrom"`
    ErrorCode     string `json:"errorCode"`
    ErrorMessage  string `json:"errorMessage"`
    Success       bool   `json:"success"`
    Value         struct {
        LastPoolDraw struct {
            LotteryDrawNum  string `json:"lotteryDrawNum"`
            LotteryDrawTime string `json:"lotteryDrawTime"`
        } `json:"lastPoolDraw"`
    } `json:"value"`
}

// GetLotteryDrawInfo 获取彩票开奖信息，优先使用API获取，失败后使用计算方式
func GetLotteryDrawInfo(lotteryCode string, scheduleCron string, apiEndpoint string) (*LotteryDrawInfo, error) {
	logger.Info("开始获取彩票[%s]开奖信息...", lotteryCode)
	
// 首先尝试从API获取
apiInfo, err := getDrawInfoFromAPI(lotteryCode, apiEndpoint)
	if err == nil && apiInfo != nil && apiInfo.CurrentDrawNum != "" {
		logger.Info("从API成功获取彩票[%s]期号信息: %s", lotteryCode, apiInfo.CurrentDrawNum)
		return apiInfo, nil
	}
	
	logger.Info("从API获取彩票[%s]期号信息失败: %v，切换到计算模式", lotteryCode, err)
	
	// API获取失败，使用计算方式
	now := time.Now()
	info := &LotteryDrawInfo{
		Code: lotteryCode,
	}

	// 解析cron表达式中的开奖日期
	drawDays, err := getDrawDaysFromCron(scheduleCron)
	if err != nil {
		return nil, err
	}

	// 根据彩票类型确定名称和期号规则
	switch lotteryCode {
	case "fc_ssq":
		info.Name = "双色球"
	case "tc_dlt":
		info.Name = "大乐透"
	default:
		return nil, fmt.Errorf("不支持的彩票类型: %s", lotteryCode)
	}

	// 计算下一个开奖日期
	drawDate, isToday := calculateNextDrawDate(now, drawDays)
	info.CurrentDrawDate = drawDate
	info.IsDrawToday = isToday

	// 生成期号
	info.CurrentDrawNum = generateDrawNumber(lotteryCode, drawDate)
	logger.Info("通过计算得到彩票[%s]期号: %s", lotteryCode, info.CurrentDrawNum)

	return info, nil
}

// getDrawInfoFromAPI 从官方API获取最新的开奖信息
func getDrawInfoFromAPI(lotteryCode string, apiEndpoint string) (*LotteryDrawInfo, error) {
if apiEndpoint == "" {
    return nil, fmt.Errorf("彩票[%s]未配置API地址", lotteryCode)
}

	
	// 设置超时时间
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
logger.Info("正在请求彩票API: %s", apiEndpoint)
req, err := http.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 添加请求头，模拟浏览器访问
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Add("Accept", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求API失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应数据失败: %v", err)
	}
	
	logger.Info("API返回数据: %s", string(body))
	
var (
    drawDate time.Time
    drawNumber string
)

switch lotteryCode {
case "tc_dlt":
    var sportResp SportLotteryResponse
    if err := json.Unmarshal(body, &sportResp); err != nil {
        return nil, fmt.Errorf("解析API响应失败: %v", err)
    }
    
    if !sportResp.Success || sportResp.ErrorCode != "0" {
        return nil, fmt.Errorf("API返回错误: %s", sportResp.ErrorMessage)
    }
    
    // 解析开奖日期
    drawDate, err = time.Parse("2006-01-02", sportResp.Value.LastPoolDraw.LotteryDrawTime)
    if err != nil {
        return nil, fmt.Errorf("解析开奖日期失败: %v", err)
    }
    drawNumber = sportResp.Value.LastPoolDraw.LotteryDrawNum
    
default:
    var apiResp LotteryAPIResponse
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return nil, fmt.Errorf("解析API响应失败: %v", err)
    }
    
    if apiResp.Code != 0 {
        return nil, fmt.Errorf("API返回错误: %s", apiResp.Message)
    }
    
    // 解析开奖日期
    drawDate, err = time.Parse("2006-01-02", apiResp.Data.DrawDate)
    if err != nil {
        return nil, fmt.Errorf("解析开奖日期失败: %v", err)
    }
    drawNumber = apiResp.Data.DrawNumber
}
	
	// 设置开奖时间为晚上8点
	drawDate = time.Date(drawDate.Year(), drawDate.Month(), drawDate.Day(), 20, 0, 0, 0, drawDate.Location())
	
	// 判断是否为今天开奖
	isToday := time.Now().Format("2006-01-02") == drawDate.Format("2006-01-02")
	
// 构造返回信息
var name string
switch lotteryCode {
case "fc_ssq":
    name = "双色球"
case "tc_dlt":
    name = "大乐透"
default:
    name = lotteryCode
}

return &LotteryDrawInfo{
    Code:            lotteryCode,
    Name:            name,
    CurrentDrawDate: drawDate,
    CurrentDrawNum:  drawNumber,
    IsDrawToday:     isToday,
}, nil
}

// getDrawDaysFromCron 从cron表达式中解析出开奖日(星期几)
// 假定cron表达式格式为 "0 0 20 * * 1,3,6" 或 "0 0 20 * * 2,4,0"
func getDrawDaysFromCron(cron string) ([]int, error) {
	parts := strings.Split(cron, " ")
	if len(parts) < 6 {
		// 可能是使用了5段式cron表达式
		if len(parts) == 5 {
			// 假设第5段是星期几，如 "0 0 20 * 1,3,6"
			dayPart := parts[4]
			return parseDaysPart(dayPart)
		}
		return nil, fmt.Errorf("无效的cron表达式: %s", cron)
	}

	// 6段式 cron 的最后一段是星期几
	dayPart := parts[5]
	return parseDaysPart(dayPart)
}

// parseDaysPart 解析cron表达式中的日期部分
func parseDaysPart(dayPart string) ([]int, error) {
	var days []int
	
	// 如果包含逗号，表示多个日期
	if strings.Contains(dayPart, ",") {
		for _, day := range strings.Split(dayPart, ",") {
			d := 0
			_, err := fmt.Sscanf(day, "%d", &d)
			if err != nil {
				return nil, fmt.Errorf("无效的开奖日: %s", day)
			}
			days = append(days, d)
		}
	} else {
		// 单个日期
		d := 0
		_, err := fmt.Sscanf(dayPart, "%d", &d)
		if err != nil {
			return nil, fmt.Errorf("无效的开奖日: %s", dayPart)
		}
		days = append(days, d)
	}

	return days, nil
}

// calculateNextDrawDate 计算下一个开奖日期
func calculateNextDrawDate(now time.Time, drawDays []int) (time.Time, bool) {
	// 获取当前是星期几（0=星期日，1=星期一...）
	currentWeekday := int(now.Weekday())
	
	// 判断今天是否是开奖日
	isToday := false
	for _, day := range drawDays {
		if day == currentWeekday {
			// 如果是开奖日，检查是否已过开奖时间（假设20:00是开奖时间）
			drawTimeToday := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
			if now.Before(drawTimeToday) {
				// 未过开奖时间，返回今天
				return drawTimeToday, true
			}
			isToday = false // 过了开奖时间
			break
		}
	}

	// 找到下一个开奖日
	daysToAdd := 7 // 最多需要加7天
	for i := 1; i <= 7; i++ {
		nextDay := (currentWeekday + i) % 7
		for _, day := range drawDays {
			if nextDay == day {
				daysToAdd = i
				break
			}
		}
		if daysToAdd < 7 {
			break
		}
	}

	nextDrawDate := now.AddDate(0, 0, daysToAdd)
	// 设置开奖时间为20:00
	nextDrawDate = time.Date(nextDrawDate.Year(), nextDrawDate.Month(), nextDrawDate.Day(), 
		20, 0, 0, 0, now.Location())
	
	return nextDrawDate, isToday
}

// generateDrawNumber 生成彩票期号
// 双色球：年份后两位 + 期号（如：23001）
// 大乐透：年份后两位 + 期号（如：23001）
func generateDrawNumber(lotteryCode string, drawDate time.Time) string {
	year := drawDate.Year() % 100 // 取年份后两位
	
	// 获取该年第一天
	firstDay := time.Date(drawDate.Year(), 1, 1, 0, 0, 0, 0, drawDate.Location())
	
	// 计算当前是一年中的第几周
	_, week := drawDate.ISOWeek()
	
	// 根据彩票类型调整期号生成逻辑
	var drawNum string
	switch lotteryCode {
	case "fc_ssq":
		// 双色球：每周开奖3次，期号从1开始
		// 简单实现：用第几周 * 3 + 星期几的调整值
		adjustment := 0
		switch drawDate.Weekday() {
		case time.Sunday:
			adjustment = 3 // 周日是当周第3次
		case time.Tuesday:
			adjustment = 1 // 周二是当周第1次
		case time.Thursday:
			adjustment = 2 // 周四是当周第2次
		}
		drawNum = fmt.Sprintf("%02d%03d", year, (week-1)*3+adjustment)
	case "tc_dlt":
		// 大乐透：每周开奖3次，期号从1开始
		// 简单实现：用第几周 * 3 + 星期几的调整值
		adjustment := 0
		switch drawDate.Weekday() {
		case time.Monday:
			adjustment = 1 // 周一是当周第1次
		case time.Wednesday:
			adjustment = 2 // 周三是当周第2次
		case time.Saturday:
			adjustment = 3 // 周六是当周第3次
		}
		drawNum = fmt.Sprintf("%02d%03d", year, (week-1)*3+adjustment)
	}
	
	return drawNum
}
