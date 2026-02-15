package middleware

import (
	"jwt-auth/utils"

	"github.com/gofiber/fiber/v2"
)

func RequireAuth(c *fiber.Ctx) error {
	token := c.Cookies("access_token")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Access token not found",
		})
	}

	claims, err := utils.ValidateToken(token, false)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired access token",
		})
	}

	c.Locals("userID", claims.UserID)
	c.Locals("email", claims.Email)

	return c.Next()
}
