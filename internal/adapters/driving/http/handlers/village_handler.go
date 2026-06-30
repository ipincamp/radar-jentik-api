package handlers

import (
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/utils"
)

type VillageHandler struct {
	villageService ports.VillageService
}

func NewVillageHandler(villageService ports.VillageService) *VillageHandler {
	return &VillageHandler{villageService: villageService}
}

// Fungsi GetAll (Untuk Dropdown pilihan desa di Flutter)
func (h *VillageHandler) GetAll(c *fiber.Ctx) error {
	// 1. Tangkap Query Parameter (Default: page 1, limit 10)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// 2. Lempar ke Service
	villages, totalData, err := h.villageService.GetPaginated(c.Context(), page, limit)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data desa", err.Error())
	}

	// 3. Susun Metadata Pagination
	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))
	meta := utils.PaginationMeta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  totalData,
		TotalPages:  totalPages,
	}

	// 4. Return Data (Akan otomatis terurut dari A-Z berkat Repo)
	return utils.Paginated(c, fiber.StatusOK, "Daftar desa berhasil diambil", villages, meta)
}
