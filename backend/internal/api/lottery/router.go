package lottery

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router) {
	group := router.Group("/lotteries")
	group.Get("/", ListLotteries)
	group.Get("/dashboard", GetGlobalDashboard)
	group.Get("/recommendations", ListAllRecommendations)
	group.Post("/draws/sync-history", SyncMultipleDraws)
	group.Get("/tickets/history", ListTicketHistory)
	group.Get("/tickets", ListAllTickets)
	group.Delete("/tickets/:ticketId", DeleteGenericTicket)
	group.Post("/tickets/:ticketId/recheck", RecheckGenericTicket)
	group.Post("/tickets/upload-image", UploadGenericTicketImage)
	group.Post("/tickets/recognize", RecognizeGenericTicket)
	group.Post("/tickets", CreateGenericTicket)
	group.Get("/:code/dashboard", GetDashboard)
	group.Get("/:code/recommendations", ListRecommendations)
	group.Get("/:code/recommendations/latest", GetLatestRecommendation)
	group.Get("/:code/recommendations/:recommendationId", GetRecommendationDetail)
	group.Delete("/:code/recommendations/:recommendationId", DeleteRecommendation)
	group.Post("/:code/recommendations/generate", GenerateRecommendation)
	group.Post("/:code/draws/sync", SyncDraws)
	group.Post("/:code/draws/sync-history", SyncDrawHistory)
	group.Get("/:code/tickets", ListTickets)
	group.Post("/:code/tickets/:ticketId/recheck", RecheckTicket)
	group.Post("/:code/tickets/upload-image", UploadTicketImage)
	group.Post("/:code/tickets/recognize", RecognizeTicket)
	group.Post("/:code/tickets", CreateTicket)
	group.Post("/:code/tickets/scan", ScanTicket)
}
