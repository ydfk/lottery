package db

import (
	lotteryModel "go-fiber-starter/internal/model/lottery"
	userModel "go-fiber-starter/internal/model/user"
)

func autoMigrate() error {
	return DB.AutoMigrate(
		&userModel.User{},
		&lotteryModel.LotteryType{},
		&lotteryModel.DrawResult{},
		&lotteryModel.DrawPrize{},
		&lotteryModel.TicketUpload{},
		&lotteryModel.Ticket{},
		&lotteryModel.TicketEntry{},
		&lotteryModel.Recommendation{},
		&lotteryModel.RecommendationEntry{},
	)
}
