package auth

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

type UserData struct {
	ID       string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Username string `json:"username" example:"alice"`
}

type UserResponse struct {
	Flag bool     `json:"flag" example:"true"`
	Code int      `json:"code" example:"200"`
	Data UserData `json:"data"`
	Time string   `json:"time" example:"2026-03-16T10:00:00Z"`
}
