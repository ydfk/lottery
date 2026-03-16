/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-09 16:37:32
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-10 16:25:20
 */
package middleware

import (
	"go-fiber-starter/internal/api/response"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	return response.Error(c, err.Error(), fiber.StatusInternalServerError)
}
