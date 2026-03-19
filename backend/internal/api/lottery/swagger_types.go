package lottery

import (
	model "go-fiber-starter/internal/model/lottery"
	lotteryService "go-fiber-starter/internal/service/lottery"
)

type ErrorResponse struct {
	Flag bool   `json:"flag" example:"false"`
	Code int    `json:"code" example:"500"`
	Msg  string `json:"msg" example:"请求失败"`
	Time string `json:"time" example:"2026-03-16T10:00:00Z"`
}

type LotteryListResponse struct {
	Flag bool                `json:"flag" example:"true"`
	Code int                 `json:"code" example:"200"`
	Data []model.LotteryType `json:"data"`
	Time string              `json:"time" example:"2026-03-16T10:00:00Z"`
}

type DashboardResponse struct {
	Flag bool                         `json:"flag" example:"true"`
	Code int                          `json:"code" example:"200"`
	Data lotteryService.DashboardData `json:"data"`
	Time string                       `json:"time" example:"2026-03-16T10:00:00Z"`
}

type RecommendationResponse struct {
	Flag bool                 `json:"flag" example:"true"`
	Code int                  `json:"code" example:"200"`
	Data model.Recommendation `json:"data"`
	Time string               `json:"time" example:"2026-03-16T10:00:00Z"`
}

type RecommendationListResponse struct {
	Flag bool                                  `json:"flag" example:"true"`
	Code int                                   `json:"code" example:"200"`
	Data []lotteryService.RecommendationDetail `json:"data"`
	Time string                                `json:"time" example:"2026-03-16T10:00:00Z"`
}

type RecommendationPageResponse struct {
	Flag bool                                    `json:"flag" example:"true"`
	Code int                                     `json:"code" example:"200"`
	Data lotteryService.RecommendationPageResult `json:"data"`
	Time string                                  `json:"time" example:"2026-03-16T10:00:00Z"`
}

type RecommendationDetailResponse struct {
	Flag bool                                `json:"flag" example:"true"`
	Code int                                 `json:"code" example:"200"`
	Data lotteryService.RecommendationDetail `json:"data"`
	Time string                              `json:"time" example:"2026-03-16T10:00:00Z"`
}

type SyncResultResponse struct {
	Flag bool                      `json:"flag" example:"true"`
	Code int                       `json:"code" example:"200"`
	Data lotteryService.SyncResult `json:"data"`
	Time string                    `json:"time" example:"2026-03-16T10:00:00Z"`
}

type BatchSyncResultResponse struct {
	Flag bool                           `json:"flag" example:"true"`
	Code int                            `json:"code" example:"200"`
	Data lotteryService.BatchSyncResult `json:"data"`
	Time string                         `json:"time" example:"2026-03-16T10:00:00Z"`
}

type TicketListResponse struct {
	Flag bool                          `json:"flag" example:"true"`
	Code int                           `json:"code" example:"200"`
	Data []lotteryService.TicketDetail `json:"data"`
	Time string                        `json:"time" example:"2026-03-16T10:00:00Z"`
}

type TicketPageResponse struct {
	Flag bool                            `json:"flag" example:"true"`
	Code int                             `json:"code" example:"200"`
	Data lotteryService.TicketPageResult `json:"data"`
	Time string                          `json:"time" example:"2026-03-16T10:00:00Z"`
}

type TicketDetailResponse struct {
	Flag bool                        `json:"flag" example:"true"`
	Code int                         `json:"code" example:"201"`
	Data lotteryService.TicketDetail `json:"data"`
	Time string                      `json:"time" example:"2026-03-16T10:00:00Z"`
}

type TicketUploadResponse struct {
	Flag bool                              `json:"flag" example:"true"`
	Code int                               `json:"code" example:"201"`
	Data lotteryService.TicketUploadDetail `json:"data"`
	Time string                            `json:"time" example:"2026-03-16T10:00:00Z"`
}

type TicketRecognitionResponse struct {
	Flag bool                                  `json:"flag" example:"true"`
	Code int                                   `json:"code" example:"200"`
	Data lotteryService.TicketRecognitionDraft `json:"data"`
	Time string                                `json:"time" example:"2026-03-16T10:00:00Z"`
}
