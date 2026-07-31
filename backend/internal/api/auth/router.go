/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-06-09 17:47:24
 * @LastEditors: ydfk
 * @LastEditTime: 2025-06-09 17:47:38
 */
package auth

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterUnProtectedRoutes(router *fiber.App) {
	grp := router.Group("/api/auth")
	grp.Post("/register", Register)
	grp.Post("/login", Login)
	grp.Post("/passkeys/login/options", BeginPasskeyLogin)
	grp.Post("/passkeys/login/verify", FinishPasskeyLogin)
}

func RegisterRoutes(router fiber.Router) {
	grp := router.Group("/auth")
	grp.Get("/profile", Profile)
	grp.Post("/passkeys/registration/options", BeginPasskeyRegistration)
	grp.Post("/passkeys/registration/verify", FinishPasskeyRegistration)
}

func RegisterAssociatedDomainRoutes(router *fiber.App) {
	router.Get("/.well-known/apple-app-site-association", AppleAppSiteAssociation)
}
