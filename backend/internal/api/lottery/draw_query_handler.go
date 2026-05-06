package lottery

import (
	"go-fiber-starter/internal/api/response"
	lotteryService "go-fiber-starter/internal/service/lottery"

	"github.com/gofiber/fiber/v2"
)

// @Summary 分页获取历史开奖记录
// @Description 返回系统已同步入库的历史开奖记录，支持按彩种筛选，并按开奖时间排序
// @Tags lottery
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认 1"
// @Param pageSize query int false "每页数量，默认 10，最大 50"
// @Param lotteryCode query string false "彩票编码，如 ssq、dlt"
// @Param sort query string false "排序，可选 latest、oldest"
// @Success 200 {object} DrawPageResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/draws/history [get]
func ListDrawHistory(c *fiber.Ctx) error {
	if _, err := currentUserID(c); err != nil {
		return err
	}

	data, err := lotteryService.QueryDrawResults(lotteryService.DrawQueryOptions{
		Page:        parseIntValue(c.Query("page"), 1),
		PageSize:    parseIntValue(c.Query("pageSize"), 10),
		LotteryCode: c.Query("lotteryCode"),
		Sort:        c.Query("sort", "latest"),
	})
	if err != nil {
		return err
	}
	return response.Success(c, data)
}
