package controllers

import (
	"fmt"
	"jwt-auth/config"
	"jwt-auth/models"
	"jwt-auth/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

var secureCookie bool = os.Getenv("ENV") == "production"

func SlowEndpoint(c *fiber.Ctx) error {
	start := time.Now()
	requestID := c.Query("id", "unknown")

	fmt.Printf("🔵 Request %s başladı: %s\n", requestID, start.Format("15:04:05.000"))

	// 3 saniye bekle (database query simülasyonu)
	//time.Sleep(2 * time.Second)

	end := time.Now()
	fmt.Printf("🟢 Request %s bitti: %s (Süre: %v)\n", requestID, end.Format("15:04:05.000"), end.Sub(start))

	return c.JSON(fiber.Map{
		"request_id": requestID,
		"started_at": start.Format("15:04:05.000"),
		"ended_at":   end.Format("15:04:05.000"),
		"duration":   end.Sub(start).String(),
		"message":    "İşlem tamamlandı",
	})
}

func Register(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if input.Email == "" || input.Password == "" || input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, password and name are required",
		})
	}

	if len(input.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters long",
		})
	}

	user := models.User{
		Email: input.Email,
		Name:  input.Name,
	}

	if err := user.HashPassword(input.Password); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Password hashing failed",
		})
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		HTTPOnly: true,
		MaxAge:   86400,
		Path:     "/",
		Secure:   secureCookie,
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}

func Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	var user models.User

	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	if err := user.CheckPassword(input.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		Secure:   secureCookie,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"user":    user,
	})
}

func GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func Logout(c *fiber.Ctx) error {

	c.ClearCookie("jwt")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout successful",
	})
}
