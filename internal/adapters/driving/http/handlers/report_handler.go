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

	// 1. Parsing Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format input tidak valid"})
	}

	// 2. Validasi Input
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. Ambil User ID dari Context (Middleware Auth)
	// Pastikan middleware Anda menyimpan claim "sub" atau "user_id" ke c.Locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: User ID tidak ditemukan"})
	}

	// 4. Panggil Service
	if err := h.service.Create(c.Context(), userID, req); err != nil {
		// Log error di sisi server untuk debugging
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan laporan"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Laporan jentik berhasil dikirim",
	})
}

func (h *ReportHandler) GetAll(c *fiber.Ctx) error {
	// Parsing Query Params (otomatis mapping ke struct berdasar tag query)
	req := ports.FindReportsRequest{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
	}

	// Ambil data dari Token (via Middleware)
	userID, _ := c.Locals("user_id").(string)
	role, _ := c.Locals("role").(string)

	req.RequestorID = userID
	req.RequestorRole = role

	resp, err := h.service.GetAll(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}

func (h *ReportHandler) Validate(c *fiber.Ctx) error {
	reportID := c.Params("id")
	var req ports.ValidateReportRequest

	// 1. Parsing Body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format input tidak valid"})
	}

	// 2. Validasi Input (Status harus 'verified' atau 'rejected')
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. Ambil ID Petugas dari Token (Verifier)
	verifierID, ok := c.Locals("user_id").(string)
	if !ok || verifierID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 4. Panggil Service
	if err := h.service.Validate(c.Context(), reportID, verifierID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memvalidasi laporan"})
	}

	return c.JSON(fiber.Map{"message": "Status laporan berhasil diperbarui"})
}

func (h *ReportHandler) GetHeatmap(c *fiber.Ctx) error {
	// 1. Parsing Query Parameters dengan Default Value
	// "p" = Power Parameter (Default 2.0)
	// "res" = Grid Resolution (Default 0.005 derajat)
	req := ports.GetHeatmapRequest{
		PowerParameter: c.QueryFloat("p", 2.0),
		GridResolution: c.QueryFloat("res", 0.005),
	}

	// 2. Panggil Service
	heatmapPoints, err := h.service.GetHeatmapData(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengkalkulasi heatmap: " + err.Error(),
		})
	}

	// 3. Return Response (JSON Array of Points)
	// Format ini mudah dikonsumsi library peta di Mobile (Google Maps / Mapbox)
	return c.JSON(fiber.Map{
		"meta": fiber.Map{
			"power_parameter": req.PowerParameter,
			"grid_resolution": req.GridResolution,
			"total_points":    len(heatmapPoints),
		},
		"data": heatmapPoints,
	})
}
