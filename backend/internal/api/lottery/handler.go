package lottery

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-fiber-starter/internal/api/response"
	lotteryService "go-fiber-starter/internal/service/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/util"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GenerateRecommendationRequest struct {
	Count int `json:"count"`
}

type SyncDrawRequest struct {
	Issue        string   `json:"issue"`
	Start        int      `json:"start"`
	Count        int      `json:"count"`
	LotteryCodes []string `json:"lotteryCodes"`
}

type RecognizeTicketRequest struct {
	UploadID string `json:"uploadId"`
	OCRText  string `json:"ocrText"`
}

type CreateTicketEntryRequest struct {
	RedNumbers  string `json:"redNumbers"`
	BlueNumbers string `json:"blueNumbers"`
}

type CreateTicketRequest struct {
	UploadID    string                     `json:"uploadId"`
	Issue       string                     `json:"issue"`
	PurchasedAt string                     `json:"purchasedAt"`
	Notes       string                     `json:"notes"`
	Entries     []CreateTicketEntryRequest `json:"entries"`
}

// @Summary 获取已启用彩票列表
// @Description 返回系统当前已加载并入库的彩票类型
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Success 200 {object} LotteryListResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/ [get]
func ListLotteries(c *fiber.Ctx) error {
	items, err := lotteryService.ListLotteryTypes()
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

// @Summary 获取彩票看板
// @Description 返回指定彩票的最新开奖、最新推荐、最近票据和统计信息
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Success 200 {object} DashboardResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/dashboard [get]
func GetDashboard(c *fiber.Ctx) error {
	data, err := lotteryService.GetDashboard(c.Params("code"))
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 获取最新推荐
// @Description 返回指定彩票最近一次生成的推荐号码
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Success 200 {object} RecommendationResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/recommendations/latest [get]
func GetLatestRecommendation(c *fiber.Ctx) error {
	data, err := lotteryService.GetLatestRecommendation(c.Params("code"))
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 生成推荐号码
// @Description 按当前彩票配置的 AI 模型和提示词生成推荐号码
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param request body GenerateRecommendationRequest false "推荐生成参数"
// @Success 200 {object} RecommendationResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/recommendations/generate [post]
func GenerateRecommendation(c *fiber.Ctx) error {
	request := GenerateRecommendationRequest{}
	if err := c.BodyParser(&request); err != nil {
		request.Count = 0
	}

	data, err := lotteryService.GenerateRecommendation(c.Context(), c.Params("code"), request.Count)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 同步当期开奖
// @Description 使用极速数据 query 接口同步当前期或指定期的开奖信息，并触发票据与推荐结算
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param request body SyncDrawRequest false "同步参数，issue 为空时同步当前期"
// @Success 200 {object} SyncResultResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/draws/sync [post]
func SyncDraws(c *fiber.Ctx) error {
	request := parseSyncDrawRequest(c)
	data, err := lotteryService.SyncLatestDraw(c.Context(), c.Params("code"), request.Issue)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 同步单种彩票历史开奖
// @Description 使用极速数据 history 接口分页同步指定彩票的历史开奖信息，默认最近 100 期
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param request body SyncDrawRequest false "历史同步参数"
// @Success 200 {object} SyncResultResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/draws/sync-history [post]
func SyncDrawHistory(c *fiber.Ctx) error {
	request := parseSyncDrawRequest(c)
	data, err := lotteryService.SyncDrawHistory(c.Context(), c.Params("code"), lotteryService.SyncOptions{
		Issue: request.Issue,
		Start: request.Start,
		Count: request.Count,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 批量同步多种彩票历史开奖
// @Description 批量调用历史开奖同步，lotteryCodes 为空时同步所有已启用彩票
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SyncDrawRequest false "批量历史同步参数"
// @Success 200 {object} BatchSyncResultResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/draws/sync-history [post]
func SyncMultipleDraws(c *fiber.Ctx) error {
	request := parseSyncDrawRequest(c)
	data, err := lotteryService.SyncMultipleDraws(c.Context(), request.LotteryCodes, lotteryService.SyncOptions{
		Issue: request.Issue,
		Start: request.Start,
		Count: request.Count,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 获取票据列表
// @Description 返回指定彩票最近录入的票据记录
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param limit query int false "返回数量，默认 20"
// @Success 200 {object} TicketListResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/tickets [get]
func ListTickets(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	items, err := lotteryService.ListTickets(c.Params("code"), limit)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

// @Summary 上传彩票原图
// @Description 仅上传并保存彩票原图，不执行 OCR 和入库
// @Tags lottery
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq"
// @Param image formData file true "彩票图片"
// @Success 201 {object} TicketUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/tickets/upload-image [post]
func UploadTicketImage(c *fiber.Ctx) error {
	file, imagePath, err := saveTicketImage(c)
	if err != nil {
		return err
	}

	data, err := lotteryService.UploadTicketImage(lotteryService.UploadTicketImageInput{
		Code:             c.Params("code"),
		ImagePath:        imagePath,
		OriginalFilename: file.Filename,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data, fiber.StatusCreated)
}

// @Summary 识别彩票内容
// @Description 基于已上传原图执行 OCR 识别，只返回识别结果，不入库
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq"
// @Param request body RecognizeTicketRequest true "识别参数"
// @Success 200 {object} TicketRecognitionResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/tickets/recognize [post]
func RecognizeTicket(c *fiber.Ctx) error {
	request := RecognizeTicketRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}

	data, err := lotteryService.RecognizeUploadedTicket(c.Context(), lotteryService.RecognizeUploadedTicketInput{
		Code:     c.Params("code"),
		UploadID: request.UploadID,
		OCRText:  request.OCRText,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 确认入库并判奖
// @Description 基于上传记录和识别结果确认票据入库，并在已开奖时自动判奖
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq"
// @Param request body CreateTicketRequest true "入库参数"
// @Success 201 {object} TicketDetailResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/tickets [post]
func CreateTicket(c *fiber.Ctx) error {
	request := CreateTicketRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}

	purchasedAt, _ := time.Parse(time.RFC3339, request.PurchasedAt)
	entries, err := parseCreateTicketEntries(request.Entries)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest)
	}

	data, err := lotteryService.CreateTicket(c.Context(), lotteryService.CreateTicketInput{
		Code:        c.Params("code"),
		UploadID:    request.UploadID,
		Issue:       request.Issue,
		PurchasedAt: purchasedAt,
		Notes:       request.Notes,
		Entries:     entries,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data, fiber.StatusCreated)
}

// @Summary 扫描彩票票据
// @Description 上传票据图片并识别号码，识别成功后自动入库并尝试判奖
// @Tags lottery
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq"
// @Param image formData file true "彩票图片"
// @Param issue formData string false "期号"
// @Param purchasedAt formData string false "购买时间，RFC3339 格式"
// @Param ocrText formData string false "OCR 降级文本"
// @Param notes formData string false "备注"
// @Success 201 {object} TicketDetailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/tickets/scan [post]
func ScanTicket(c *fiber.Ctx) error {
	_, imagePath, err := saveTicketImage(c)
	if err != nil {
		return err
	}

	purchasedAt, _ := time.Parse(time.RFC3339, c.FormValue("purchasedAt"))
	data, err := lotteryService.ScanTicket(c.Context(), lotteryService.ScanTicketInput{
		Code:        c.Params("code"),
		Issue:       c.FormValue("issue"),
		ImagePath:   imagePath,
		OCRText:     c.FormValue("ocrText"),
		PurchasedAt: purchasedAt,
		Notes:       c.FormValue("notes"),
	})
	if err != nil {
		return err
	}
	return response.Success(c, data, fiber.StatusCreated)
}

func saveTicketImage(c *fiber.Ctx) (*multipart.FileHeader, string, error) {
	file, err := c.FormFile("image")
	if err != nil {
		return nil, "", response.Error(c, "请上传彩票图片", fiber.StatusBadRequest)
	}

	imagePath := filepath.Join(
		config.Current.Storage.UploadDir,
		"tickets",
		time.Now().Format("20060102"),
		uuid.NewString()+filepath.Ext(file.Filename),
	)
	if err := util.EnsureDir(imagePath); err != nil {
		return nil, "", err
	}
	if err := c.SaveFile(file, imagePath); err != nil {
		return nil, "", err
	}
	return file, imagePath, nil
}

func parseCreateTicketEntries(items []CreateTicketEntryRequest) ([]lotteryService.ParsedEntry, error) {
	if len(items) == 0 {
		return nil, nil
	}

	result := make([]lotteryService.ParsedEntry, 0, len(items))
	for _, item := range items {
		if item.RedNumbers == "" || item.BlueNumbers == "" {
			return nil, fmt.Errorf("每注号码都需要包含 redNumbers 和 blueNumbers")
		}
		result = append(result, lotteryService.ParsedEntry{
			Red:  parseCSVValues(item.RedNumbers),
			Blue: parseCSVValues(item.BlueNumbers),
		})
	}
	return result, nil
}

func parseCSVValues(value string) []int {
	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		number := parseIntValue(strings.TrimSpace(part), -1)
		if number >= 0 {
			result = append(result, number)
		}
	}
	return result
}

func parseSyncDrawRequest(c *fiber.Ctx) SyncDrawRequest {
	request := SyncDrawRequest{
		Issue: c.Query("issue"),
		Start: parseIntValue(c.Query("start"), 0),
		Count: parseIntValue(c.Query("count"), 100),
	}
	if err := c.BodyParser(&request); err != nil {
		return request
	}
	if request.Count <= 0 {
		request.Count = 100
	}
	if request.Start < 0 {
		request.Start = 0
	}
	return request
}

func parseIntValue(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
