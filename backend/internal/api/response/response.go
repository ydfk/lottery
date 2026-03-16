package response

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Flag bool        `json:"flag"`
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"msg,omitempty"`
	Time string      `json:"time"`
}

func Success(c *fiber.Ctx, data interface{}, code ...int) error {
	statusCode := fiber.StatusOK
	if len(code) > 0 {
		statusCode = code[0]
	}

	return c.Status(statusCode).JSON(Response{
		Flag: true,
		Code: statusCode,
		Data: data,
		Time: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func Error(c *fiber.Ctx, msg string, code ...int) error {
	statusCode := fiber.StatusInternalServerError
	if len(code) > 0 {
		statusCode = code[0]
	}
	return c.Status(statusCode).JSON(Response{
		Flag: false,
		Code: statusCode,
		Msg:  msg,
		Time: time.Now().UTC().Format(time.RFC3339Nano),
	})
}
