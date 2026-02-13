package middleware

import (
	"jwt-auth/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func RequireAuth(c *fiber.Ctx) error {
	var token string

	token = c.Cookies("jwt")

	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token = tokenParts[1]
			}
		}
	}

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token not found",
		})
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	c.Locals("userID", claims.UserID)
	c.Locals("email", claims.Email)

	return c.Next()

}
