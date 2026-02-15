package routes

import (
	"jwt-auth/controllers"
	"jwt-auth/middleware" // Limiter burada

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(app *fiber.App) {
	api := app.Group("/api/auth")

	// public routes
	api.Post("/register", middleware.AuthLimiter(), controllers.Register)
	api.Post("/login", middleware.AuthLimiter(), controllers.Login)
	api.Post("/refresh", middleware.AuthLimiter(), controllers.Refresh)

	// private routes
	api.Get("/profile", middleware.RequireAuth, middleware.ApiLimiter(), controllers.GetProfile)
	api.Post("/logout", middleware.RequireAuth, controllers.Logout)
	api.Post("/logout-all", middleware.RequireAuth, controllers.LogoutAll)
}
