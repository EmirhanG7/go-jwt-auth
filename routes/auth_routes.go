package routes

import (
	"jwt-auth/controllers"
	"jwt-auth/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")

	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)

	auth.Get("/profile", middleware.RequireAuth, controllers.GetProfile)
	auth.Post("/logout", middleware.RequireAuth, controllers.Logout)

	auth.Get("/test-concurrent", controllers.SlowEndpoint)

}
