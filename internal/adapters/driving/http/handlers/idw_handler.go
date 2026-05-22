package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type IDWHandler struct {
	idwService ports.IDWService
}

func NewIDWHandler(idwService ports.IDWService) *IDWHandler {
	return &IDWHandler{idwService: idwService}
}

func (h *IDWHandler) Calculate(c *fiber.Ctx) error {
	var req domain.IDWRequest

	// Parsing body request JSON dari Flutter
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format payload tidak valid",
		})
	}

	// Validasi dasar
	if len(req.Samples) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Titik sampel tidak boleh kosong",
		})
	}

	// Default power parameter untuk IDW jika tidak dikirim
	if req.Power <= 0 {
		req.Power = 2.0
	}
	// Default resolution
	if req.Resolution <= 0 {
		req.Resolution = 0.001
	}

	// Panggil Service IDW
	gridResult, err := h.idwService.CalculateIDWGrid(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghitung estimasi IDW",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Berhasil menghitung IDW",
		"data":    gridResult,
	})
}
