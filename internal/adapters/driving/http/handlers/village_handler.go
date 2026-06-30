package handlers

import (
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
	villages, err := h.villageService.GetAllVillages(c.Context())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data desa", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Data desa berhasil diambil", villages)
}
