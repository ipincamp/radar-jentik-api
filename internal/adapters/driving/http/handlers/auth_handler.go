package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type AuthHandler struct {
	authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// DTO untuk Registrasi
type RegisterRequest struct {
	FullName  string `json:"full_name" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required,min=6"`
	Role      string `json:"role" validate:"required,oneof=cadre officer"`
	VillageID string `json:"village_id" validate:"required"` // UUID Desa
}

// DTO untuk Login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Fungsi Register (Buat Akun)
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Format request tidak valid",
			"details": err.Error(),
		})
	}

	user := &domain.User{
		FullName:  req.FullName,
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		VillageID: req.VillageID,
	}

	if err := h.authService.Register(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Gagal mendaftarkan pengguna",
			"details": err.Error(),
		})
	}

	// Jangan kembalikan password di response
	user.Password = ""

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Registrasi akun berhasil",
		"data":    user,
	})
}

// Fungsi Login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	token, role, err := h.authService.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Username atau password salah",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login berhasil",
		"token":   token,
		"role":    role,
	})
}

// Fungsi Logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Karena kita menggunakan JWT Stateless, logout biasanya dilakukan
	// dengan cara menghapus token di sisi aplikasi Flutter.
	// Endpoint ini hanya sebagai konfirmasi/formalitas.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout berhasil",
	})
}

// Fungsi List Users (Untuk halaman Manajemen Kader oleh Petugas)
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	// Pastikan fungsi GetAllUsers ada di auth_service.go Anda
	users, err := h.authService.GetAllUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data daftar kader",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": users,
	})
}
