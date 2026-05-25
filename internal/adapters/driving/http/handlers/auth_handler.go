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
	VillageID string `json:"village_id" validate:"required"`
}

// DTO untuk Login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// DTO untuk Update User
type UpdateUserRequest struct {
	FullName  string `json:"full_name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	VillageID string `json:"village_id"`
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

// REVISI FUNGSI READ (LIST USERS)
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.authService.GetAllUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data daftar kader"})
	}

	var response []fiber.Map
	for _, u := range users {
		// ===============================================
		// FILTER: Lewati / Jangan masukkan jika bukan kader
		// ===============================================
		if u.Role != "cadre" {
			continue
		}

		villageName := ""
		if u.Village != nil {
			villageName = u.Village.Name
		}
		response = append(response, fiber.Map{
			"id":           u.ID,
			"full_name":    u.FullName,
			"username":     u.Username,
			"role":         u.Role,
			"village_id":   u.VillageID,
			"village_name": villageName,
		})
	}

	// Mencegah return null jika list kader masih kosong murni
	if response == nil {
		response = []fiber.Map{}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": response})
}

// FUNGSI CREATE USER (PETUGAS MEMBUAT KADER BARU)
func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	req := new(RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	user := &domain.User{
		FullName:  req.FullName,
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		VillageID: req.VillageID,
	}

	if err := h.authService.CreateUser(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Akun Kader berhasil dibuat"})
}

// FUNGSI UPDATE USER
func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	req := new(UpdateUserRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	user := &domain.User{
		FullName:  req.FullName,
		Username:  req.Username,
		Password:  req.Password,
		VillageID: req.VillageID,
	}

	if err := h.authService.UpdateUser(c.Context(), id, user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Data Kader berhasil diperbarui"})
}

// FUNGSI DELETE USER
func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.authService.DeleteUser(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus akun kader"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Akun Kader berhasil dihapus"})
}
