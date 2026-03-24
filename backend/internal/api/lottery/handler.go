package lottery

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-fiber-starter/internal/api/response"
	coreService "go-fiber-starter/internal/service"
	lotteryService "go-fiber-starter/internal/service/lottery"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/util"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
	RedNumbers   string `json:"redNumbers"`
	BlueNumbers  string `json:"blueNumbers"`
	Multiple     int    `json:"multiple"`
	IsAdditional bool   `json:"isAdditional"`
}

type CreateTicketRequest struct {
	LotteryCode      string                     `json:"lotteryCode"`
	UploadID         string                     `json:"uploadId"`
	RecommendationID string                     `json:"recommendationId"`
	Issue            string                     `json:"issue"`
	DrawDate         string                     `json:"drawDate"`
	PurchasedAt      string                     `json:"purchasedAt"`
	CostAmount       float64                    `json:"costAmount"`
	Notes            string                     `json:"notes"`
	Entries          []CreateTicketEntryRequest `json:"entries"`
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
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.GetDashboard(c.Params("code"), userID)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 获取全局看板
// @Description 返回当前账户全部已录入票据的汇总统计
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DashboardResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/dashboard [get]
func GetGlobalDashboard(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.GetGlobalDashboard(userID)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 获取推荐列表
// @Description 返回指定彩票最近生成的推荐记录列表，包含是否已记录购买票据
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param limit query int false "返回数量，默认 20"
// @Success 200 {object} RecommendationListResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/recommendations [get]
func ListRecommendations(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	data, err := lotteryService.ListRecommendations(c.Params("code"), limit, userID)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 获取全部推荐列表
// @Description 返回全部彩票推荐记录，支持按彩种、开奖状态筛选，并按时间或金额排序，适合移动端动态加载
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认 1"
// @Param pageSize query int false "每页数量，默认 10，最大 50"
// @Param lotteryCode query string false "彩票编码，如 ssq、dlt"
// @Param status query string false "状态，可选 pending、won、not_won"
// @Param sort query string false "排序，可选 latest、oldest、draw_latest、draw_oldest、prize_high"
// @Success 200 {object} RecommendationPageResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/recommendations [get]
func ListAllRecommendations(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.QueryRecommendations(lotteryService.RecommendationQueryOptions{
		UserID:      userID,
		Page:        parseIntValue(c.Query("page"), 1),
		PageSize:    parseIntValue(c.Query("pageSize"), 10),
		LotteryCode: c.Query("lotteryCode"),
		Status:      c.Query("status"),
		Sort:        c.Query("sort", "latest"),
	})
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
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.GetLatestRecommendation(c.Params("code"), userID)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 获取推荐详情
// @Description 返回指定推荐记录的完整号码、命中情况和购买关联信息
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param recommendationId path string true "推荐记录 ID"
// @Success 200 {object} RecommendationDetailResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/recommendations/{recommendationId} [get]
func GetRecommendationDetail(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.GetRecommendationDetail(c.Params("code"), c.Params("recommendationId"), userID)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 删除推荐记录
// @Description 删除推荐记录及推荐明细，并解除与购买记录的关联，不删除已录入票据
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param recommendationId path string true "推荐记录 ID"
// @Success 200 {object} DeleteResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/recommendations/{recommendationId} [delete]
func DeleteRecommendation(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	if err := lotteryService.DeleteRecommendation(c.Params("code"), c.Params("recommendationId"), userID); err != nil {
		return err
	}
	return response.Success(c, fiber.Map{"deleted": true})
}

// @Summary 生成推荐号码
// @Description 按当前彩票配置的 AI 模型和提示词生成推荐号码
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Success 200 {object} RecommendationResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/recommendations/generate [post]
func GenerateRecommendation(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.GenerateRecommendation(c.Context(), c.Params("code"), 0, userID)
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
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	items, err := lotteryService.ListTickets(c.Params("code"), limit, userID)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

// @Summary 获取全部票据列表
// @Description 返回最近录入的全部票据记录，适合在不预先区分彩种的记录页使用
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param limit query int false "返回数量，默认 20"
// @Success 200 {object} TicketListResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets [get]
func ListAllTickets(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	items, err := lotteryService.ListAllTickets(limit, userID)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

// @Summary 分页获取历史票据
// @Description 支持按彩种、中奖状态筛选，并按时间或金额排序，适合移动端历史列表动态加载
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认 1"
// @Param pageSize query int false "每页数量，默认 10，最大 50"
// @Param lotteryCode query string false "彩票编码，如 ssq、dlt"
// @Param status query string false "状态，可选 pending、won、not_won"
// @Param sort query string false "排序，可选 latest、oldest、prize_high、cost_high"
// @Success 200 {object} TicketPageResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets/history [get]
func ListTicketHistory(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.QueryAllTickets(lotteryService.TicketQueryOptions{
		UserID:      userID,
		Page:        parseIntValue(c.Query("page"), 1),
		PageSize:    parseIntValue(c.Query("pageSize"), 10),
		LotteryCode: c.Query("lotteryCode"),
		Status:      c.Query("status"),
		Sort:        c.Query("sort", "latest"),
	})
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 重新判奖
// @Description 按当前期号重新同步开奖并再次判奖，适合补录历史开奖后修正票据状态
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "彩票编码，如 ssq、dlt"
// @Param ticketId path string true "票据 ID"
// @Success 200 {object} TicketDetailResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/{code}/tickets/{ticketId}/recheck [post]
func RecheckTicket(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.RecheckTicket(c.Context(), c.Params("ticketId"), c.Params("code"), userID)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 通用重新判奖
// @Description 按票据自身的彩种与期号重新同步开奖并再次判奖
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ticketId path string true "票据 ID"
// @Success 200 {object} TicketDetailResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets/{ticketId}/recheck [post]
func RecheckGenericTicket(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	data, err := lotteryService.RecheckTicket(c.Context(), c.Params("ticketId"), "", userID)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 删除票据记录
// @Description 删除票据、票据明细，以及关联的上传记录和可清理的原图文件
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param ticketId path string true "票据 ID"
// @Success 200 {object} DeleteResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets/{ticketId} [delete]
func DeleteGenericTicket(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	if err := lotteryService.DeleteTicket(c.Params("ticketId"), userID); err != nil {
		return err
	}
	return response.Success(c, fiber.Map{"deleted": true})
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
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	file, imagePath, err := saveTicketImage(c)
	if err != nil {
		return err
	}

	data, err := lotteryService.UploadTicketImage(lotteryService.UploadTicketImageInput{
		UserID:           userID,
		Code:             c.Params("code"),
		ImagePath:        imagePath,
		OriginalFilename: file.Filename,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data, fiber.StatusCreated)
}

// @Summary 上传通用彩票原图
// @Description 上传时不要求先知道彩票类型，图片会在识别阶段自动判断彩种
// @Tags lottery
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param image formData file true "彩票图片"
// @Success 201 {object} TicketUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets/upload-image [post]
func UploadGenericTicketImage(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	file, imagePath, err := saveTicketImage(c)
	if err != nil {
		return err
	}

	data, err := lotteryService.UploadTicketImage(lotteryService.UploadTicketImageInput{
		UserID:           userID,
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
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	request := RecognizeTicketRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}

	data, err := lotteryService.RecognizeUploadedTicket(c.Context(), lotteryService.RecognizeUploadedTicketInput{
		UserID:   userID,
		Code:     c.Params("code"),
		UploadID: request.UploadID,
		OCRText:  request.OCRText,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

// @Summary 识别通用彩票内容
// @Description 基于已上传原图执行 OCR 识别，并自动判断彩票类型
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RecognizeTicketRequest true "识别参数"
// @Success 200 {object} TicketRecognitionResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets/recognize [post]
func RecognizeGenericTicket(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	request := RecognizeTicketRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}

	data, err := lotteryService.RecognizeUploadedTicket(c.Context(), lotteryService.RecognizeUploadedTicketInput{
		UserID:   userID,
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
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	request := CreateTicketRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}
	if request.UploadID == "" {
		return response.Error(c, "请先上传彩票图片", fiber.StatusBadRequest)
	}

	purchasedAt, err := parseOptionalTime(request.PurchasedAt)
	if err != nil {
		return response.Error(c, "购买时间格式不正确，应为 RFC3339", fiber.StatusBadRequest)
	}
	drawDate, err := parseOptionalDate(request.DrawDate)
	if err != nil {
		return response.Error(c, "开奖日期格式不正确，应为 YYYY-MM-DD", fiber.StatusBadRequest)
	}
	entries, err := parseCreateTicketEntries(request.Entries)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest)
	}

	data, err := lotteryService.CreateTicket(c.Context(), lotteryService.CreateTicketInput{
		UserID:           userID,
		Code:             firstNonEmpty(c.Params("code"), request.LotteryCode),
		UploadID:         request.UploadID,
		RecommendationID: request.RecommendationID,
		Issue:            request.Issue,
		DrawDate:         drawDate,
		PurchasedAt:      purchasedAt,
		CostAmount:       request.CostAmount,
		Notes:            request.Notes,
		Entries:          entries,
	})
	if err != nil {
		if errors.Is(err, lotteryService.ErrDuplicateTicket) {
			return response.Error(c, err.Error(), fiber.StatusConflict)
		}
		return err
	}
	return response.Success(c, data, fiber.StatusCreated)
}

// @Summary 通用票据入库并判奖
// @Description 基于上传记录和识别结果确认票据入库，彩种由识别结果或推荐记录自动决定
// @Tags lottery
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTicketRequest true "入库参数"
// @Success 201 {object} TicketDetailResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets [post]
func CreateGenericTicket(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	request := CreateTicketRequest{}
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, "参数不正确", fiber.StatusBadRequest)
	}
	if request.UploadID == "" {
		return response.Error(c, "请先上传彩票图片", fiber.StatusBadRequest)
	}

	purchasedAt, err := parseOptionalTime(request.PurchasedAt)
	if err != nil {
		return response.Error(c, "购买时间格式不正确，应为 RFC3339", fiber.StatusBadRequest)
	}
	drawDate, err := parseOptionalDate(request.DrawDate)
	if err != nil {
		return response.Error(c, "开奖日期格式不正确，应为 YYYY-MM-DD", fiber.StatusBadRequest)
	}
	entries, err := parseCreateTicketEntries(request.Entries)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusBadRequest)
	}

	data, err := lotteryService.CreateTicket(c.Context(), lotteryService.CreateTicketInput{
		UserID:           userID,
		Code:             request.LotteryCode,
		UploadID:         request.UploadID,
		RecommendationID: request.RecommendationID,
		Issue:            request.Issue,
		DrawDate:         drawDate,
		PurchasedAt:      purchasedAt,
		CostAmount:       request.CostAmount,
		Notes:            request.Notes,
		Entries:          entries,
	})
	if err != nil {
		if errors.Is(err, lotteryService.ErrDuplicateTicket) {
			return response.Error(c, err.Error(), fiber.StatusConflict)
		}
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
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	_, imagePath, err := saveTicketImage(c)
	if err != nil {
		return err
	}

	purchasedAt, err := parseOptionalTime(c.FormValue("purchasedAt"))
	if err != nil {
		return response.Error(c, "购买时间格式不正确，应为 RFC3339", fiber.StatusBadRequest)
	}
	data, err := lotteryService.ScanTicket(c.Context(), lotteryService.ScanTicketInput{
		UserID:      userID,
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
			Red:          parseCSVValues(item.RedNumbers),
			Blue:         parseCSVValues(item.BlueNumbers),
			Multiple:     item.Multiple,
			IsAdditional: item.IsAdditional,
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

func firstNonEmpty(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func parseSyncDrawRequest(c *fiber.Ctx) SyncDrawRequest {
	request := SyncDrawRequest{
		Issue: c.Query("issue"),
		Start: parseIntValue(c.Query("start"), 0),
		Count: parseIntValue(c.Query("count"), 100),
	}

	if len(c.Body()) > 0 {
		parsed := SyncDrawRequest{}
		if err := json.Unmarshal(c.Body(), &parsed); err == nil {
			if parsed.Issue != "" {
				request.Issue = parsed.Issue
			}
			if parsed.Start != 0 {
				request.Start = parsed.Start
			}
			if parsed.Count != 0 {
				request.Count = parsed.Count
			}
			if len(parsed.LotteryCodes) > 0 {
				request.LotteryCodes = parsed.LotteryCodes
			}
		} else if err := c.BodyParser(&request); err != nil {
			return request
		}
	} else if err := c.BodyParser(&request); err != nil {
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

func currentUserID(c *fiber.Ctx) (string, error) {
	user, err := coreService.CurrentUser(c)
	if err != nil {
		return "", err
	}
	return user.Id.String(), nil
}

func parseOptionalTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}

func parseOptionalDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}
