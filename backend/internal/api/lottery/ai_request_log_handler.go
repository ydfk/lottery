package lottery

import (
	"errors"
	"strconv"
	"strings"

	"go-fiber-starter/internal/api/response"
	lotteryService "go-fiber-starter/internal/service/lottery"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ListAIRequestLogs godoc
// @Summary 查看当前账号的 AI 请求日志
// @Tags lottery
// @Security BearerAuth
// @Param page query int false "页码"
// @Param purpose query string false "调用类型" Enums(recommendation,ocr)
// @Success 200 {object} AIRequestLogPageResponse
// @Router /lotteries/ai-request-logs [get]
func ListAIRequestLogs(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusUnauthorized)
	}
	purpose := strings.TrimSpace(c.Query("purpose"))
	if purpose != "" && purpose != "recommendation" && purpose != "ocr" {
		return response.Error(c, "无效的请求类型", fiber.StatusBadRequest)
	}
	logs, err := lotteryService.ListAIRequestLogs(userID, c.QueryInt("page", 1), purpose)
	if err != nil {
		return response.Error(c, "查询 AI 请求日志失败")
	}
	c.Set("Cache-Control", "no-store")
	return response.Success(c, logs)
}

// GetAIRequestLog godoc
// @Summary 查看当前账号单次 AI 请求的详细日志
// @Tags lottery
// @Security BearerAuth
// @Param id path int true "日志 ID"
// @Success 200 {object} AIRequestLogDetailResponse
// @Router /lotteries/ai-request-logs/{id} [get]
func GetAIRequestLog(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return response.Error(c, err.Error(), fiber.StatusUnauthorized)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 63)
	if err != nil || id == 0 {
		return response.Error(c, "无效的日志 ID", fiber.StatusBadRequest)
	}
	entry, err := lotteryService.GetAIRequestLog(userID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.Error(c, "请求日志不存在", fiber.StatusNotFound)
	}
	if err != nil {
		return response.Error(c, "读取 AI 请求日志失败")
	}
	c.Set("Cache-Control", "no-store")
	return response.Success(c, entry)
}
