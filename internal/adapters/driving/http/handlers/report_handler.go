package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type ReportHandler struct {
	service   ports.ReportService
	validator *validator.Validate
}

func NewReportHandler(s ports.ReportService) *ReportHandler {
	return &ReportHandler{
		service:   s,
		validator: validator.New(),
	}
}

func (h *ReportHandler) Create(c *fiber.Ctx) error {
	var req ports.CreateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Ambil User ID dari Token (disimpan di Locals oleh middleware auth)
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if err := h.service.Create(c.Context(), userID, req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan laporan"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Laporan berhasil dikirim"})
}
