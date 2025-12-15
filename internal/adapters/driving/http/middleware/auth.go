package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/pkg/auth"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
)

// Protected adalah middleware factory yang mengembalikan handler Fiber
func Protected(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil Header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token tidak ditemukan",
			})
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Format token salah (Gunakan 'Bearer <token>')",
			})
		}

		tokenString := parts[1]

		// 3. Inisialisasi Token Manager (Helper)
		// Gunakan config yang disuntikkan untuk membuat instance validator
		tokenManager := auth.NewTokenManager(cfg)

		// 4. Validasi Token
		userID, role, err := tokenManager.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: " + err.Error(),
			})
		}

		// 5. Simpan User ID dan Role ke Context (Locals)
		// Ini adalah kunci agar handler selanjutnya bisa mengenali user
		c.Locals("user_id", userID)
		c.Locals("role", role)

		// 6. Lanjut ke Handler berikutnya
		return c.Next()
	}
}
