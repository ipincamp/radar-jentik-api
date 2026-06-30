package handlers

import (
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/utils"
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
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	user := &domain.User{
		FullName:  req.FullName,
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		VillageID: req.VillageID,
	}

	if err := h.authService.Register(c.Context(), user); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mendaftarkan pengguna", err.Error())
	}

	// Jangan kembalikan password di response
	user.Password = ""

	return utils.Success(c, fiber.StatusCreated, "Registrasi akun berhasil", user)
}

// Fungsi Login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	token, role, err := h.authService.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "Username atau password salah", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"token": token,
		"role":  role,
	})
}

// Fungsi Logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Karena kita menggunakan JWT Stateless, logout biasanya dilakukan
	// dengan cara menghapus token di sisi aplikasi Flutter.
	// Endpoint ini hanya sebagai konfirmasi/formalitas.
	return utils.Success(c, fiber.StatusOK, "Logout berhasil", nil)
}

// Fungsi List Users (Hanya untuk role officer)
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	// 1. Tangkap parameter dari URL (Default: halaman 1, limit 10 per halaman)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// 2. Lempar page dan limit ke Service & Repository
	// (Anda harus menyesuaikan repo Anda untuk memakai gorm .Offset() dan .Limit())
	users, totalData, err := h.authService.GetPaginatedUsers(c.Context(), page, limit)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data", err.Error())
	}

	// 3. Hitung total halaman (Total Data dibagi Limit, bulatkan ke atas)
	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))

	// 4. Susun Meta
	meta := utils.PaginationMeta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  totalData,
		TotalPages:  totalPages,
	}

	// 5. Kembalikan Response Paginated
	return utils.Paginated(c, fiber.StatusOK, "Berhasil mengambil data kader", users, meta)

	/*
		users, err := h.authService.GetAllUsers(c.Context())
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data daftar kader", err.Error())
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

		return utils.Success(c, fiber.StatusOK, "Data daftar kader berhasil diambil", response)
	*/
}

// Fungsi Create User (Hanya untuk role officer)
func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	req := new(RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	user := &domain.User{
		FullName:  req.FullName,
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		VillageID: req.VillageID,
	}

	if err := h.authService.CreateUser(c.Context(), user); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mendaftarkan pengguna", err.Error())
	}

	return utils.Success(c, fiber.StatusCreated, "Akun Kader berhasil dibuat", user)
}

// Fungsi Update User (Hanya untuk role officer)
func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	req := new(UpdateUserRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	user := &domain.User{
		FullName:  req.FullName,
		Username:  req.Username,
		Password:  req.Password,
		VillageID: req.VillageID,
	}

	if err := h.authService.UpdateUser(c.Context(), id, user); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal memperbarui data kader", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Data Kader berhasil diperbarui", user)
}

// Fungsi Delete User (Hanya untuk role officer)
func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.authService.DeleteUser(c.Context(), id); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal menghapus akun kader", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Akun Kader berhasil dihapus", nil)
}
