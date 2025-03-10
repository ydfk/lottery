// filepath: d:\1.Development\4.gitea\lottery\backend\internal\pkg\draw\draw.go
package draw

import (
	"encoding/json"
	"fmt"
	"io"
	"lottery-backend/internal/pkg/logger"
	"net/http"
	"time"
)

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
	DataFrom     string `json:"dataFrom"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
	Success      bool   `json:"success"`
	Value        struct {
		LastPoolDraw struct {
			LotteryDrawNum  string `json:"lotteryDrawNum"`
			LotteryDrawTime string `json:"lotteryDrawTime"`
		} `json:"lastPoolDraw"`
	} `json:"value"`
}

// GetLotteryDrawInfo 获取彩票开奖信息，从API获取
func GetLotteryDrawInfo(lotteryCode string, scheduleCron string, apiEndpoint string) (*LotteryDrawInfo, error) {
	logger.Info("开始获取彩票[%s]开奖信息...", lotteryCode)

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
		drawDate   time.Time
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
