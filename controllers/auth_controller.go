package controllers

import (
	"jwt-auth/config"
	"jwt-auth/models"
	"jwt-auth/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

var secureCookie bool = os.Getenv("ENV") == "production"

func setCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   secureCookie,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   15 * 60,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   secureCookie,
		SameSite: "Lax",
		Path:     "/api/auth/refresh",
		MaxAge:   7 * 24 * 3600,
	})
}

func clearCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   secureCookie,
		SameSite: "Lax",
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   secureCookie,
		SameSite: "Lax",
		Path:     "/api/auth/refresh",
	})
}

// validate structs
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=32"`
	Name     string `json:"name" validate:"required,min=2"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register godoc
// @Summary Yeni Kullanıcı Kaydı
// @Description Kullanıcıyı sisteme kaydeder.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterInput true "Kayıt Bilgileri"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Router /register [post]
func Register(c *fiber.Ctx) error {
	var input RegisterInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if errors := utils.ValidateStruct(input); errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	user := models.User{Email: input.Email, Name: input.Name}
	if err := user.HashPassword(input.Password); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Hashing failed"})
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already exists"})
	}

	accessToken, _ := utils.GenerateAccessToken(user.ID, user.Email)
	refreshToken, _ := utils.GenerateRefreshToken(user.ID)

	config.DB.Create(&models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})

	setCookies(c, accessToken, refreshToken)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

// Login godoc
// @Summary Kullanıcı Girişi
// @Description Email ve şifre ile giriş yapar, Access ve Refresh token döner.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginInput true "Login Bilgileri"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /login [post]
func Login(c *fiber.Ctx) error {
	var input LoginInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if errors := utils.ValidateStruct(input); errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := user.CheckPassword(input.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	accessToken, _ := utils.GenerateAccessToken(user.ID, user.Email)
	refreshToken, _ := utils.GenerateRefreshToken(user.ID)

	config.DB.Create(&models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})

	setCookies(c, accessToken, refreshToken)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"user":    user,
	})
}

// Refresh godoc
// @Summary Token Yenileme (Refresh Token Rotation)
// @Description HttpOnly Cookie içindeki Refresh Token'ı kullanarak yeni bir Access Token alır. Aynı zamanda Refresh Token'ı da yeniler (Rotation).
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Yeni Access Token döner"
// @Failure 401 {object} map[string]interface{} "Token bulunamadı veya süresi dolmuş"
// @Failure 403 {object} map[string]interface{} "Güvenlik İhlali: Token Reuse Detected!"
// @Router /refresh [post]
func Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh token not found"})
	}

	claims, err := utils.ValidateToken(refreshToken, true)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	tx := config.DB.Begin()

	var storedToken models.RefreshToken
	if err := tx.Where("token = ?", refreshToken).First(&storedToken).Error; err != nil {
		tx.Rollback()
		config.DB.Where("user_id = ?", claims.UserID).Delete(&models.RefreshToken{})
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Security alert: Token reuse detected"})
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		tx.Delete(&storedToken)
		tx.Commit()
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh token expired"})
	}

	tx.Delete(&storedToken)

	newAccessToken, _ := utils.GenerateAccessToken(claims.UserID, "")
	newRefreshToken, _ := utils.GenerateRefreshToken(claims.UserID)

	tx.Create(&models.RefreshToken{
		UserID:    claims.UserID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	})

	tx.Commit()

	setCookies(c, newAccessToken, newRefreshToken)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Token refreshed"})
}

// Logout godoc
// @Summary Çıkış Yap (Logout)
// @Description Geçerli oturumu kapatır. Veritabanından mevcut refresh token'ı siler ve tarayıcıdaki cookie'leri temizler.
// @Tags Auth
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Başarıyla çıkış yapıldı"
// @Router /logout [post]
func Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		config.DB.Where("token = ?", refreshToken).Delete(&models.RefreshToken{})
	}

	clearCookies(c)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout successful",
	})
}

// LogoutAll godoc
// @Summary Tüm Cihazlardan Çıkış Yap
// @Description Kullanıcının veritabanındaki TÜM refresh tokenlarını siler (Mobil, Web, Tablet hepsi düşer).
// @Tags Auth
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /logout-all [post]
func LogoutAll(c *fiber.Ctx) error {
	userIDLocal := c.Locals("userID")
	if userIDLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID := userIDLocal.(uint)

	if err := config.DB.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	clearCookies(c)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out from all devices",
	})
}

// GetProfile godoc
// @Summary Kullanıcı Profili
// @Description Giriş yapmış kullanıcının bilgilerini döner.
// @Tags User
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.User
// @Router /profile [get]
func GetProfile(c *fiber.Ctx) error {
	userIDLocal := c.Locals("userID")
	if userIDLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	userID := userIDLocal.(uint)

	var user models.User

	result := config.DB.First(&user, userID)

	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}
