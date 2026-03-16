package auth

import model "go-fiber-starter/internal/model/user"

type ErrorResponse struct {
	Flag bool   `json:"flag" example:"false"`
	Code int    `json:"code" example:"500"`
	Msg  string `json:"msg" example:"请求失败"`
	Time string `json:"time" example:"2026-03-16T10:00:00Z"`
}

type LoginData struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`
}

type LoginResponse struct {
	Flag bool      `json:"flag" example:"true"`
	Code int       `json:"code" example:"200"`
	Data LoginData `json:"data"`
	Time string    `json:"time" example:"2026-03-16T10:00:00Z"`
}

type UserResponse struct {
	Flag bool       `json:"flag" example:"true"`
	Code int        `json:"code" example:"200"`
	Data model.User `json:"data"`
	Time string     `json:"time" example:"2026-03-16T10:00:00Z"`
}
