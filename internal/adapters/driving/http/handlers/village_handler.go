package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type VillageHandler struct {
	villageService ports.VillageService // Pastikan Anda sudah membuat interface ini di layer ports
}

func NewVillageHandler(villageService ports.VillageService) *VillageHandler {
	return &VillageHandler{villageService: villageService}
}

// Fungsi GetAll (Untuk Dropdown pilihan desa di Flutter)
func (h *VillageHandler) GetAll(c *fiber.Ctx) error {
	villages, err := h.villageService.GetAllVillages(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Gagal mengambil data desa",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data desa berhasil diambil",
		"data":    villages,
	})
}
