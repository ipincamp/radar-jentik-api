package handlers

import (
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type InspectionReportHandler struct {
	service ports.InspectionReportService
}

func NewInspectionReportHandler(service ports.InspectionReportService) *InspectionReportHandler {
	return &InspectionReportHandler{service: service}
}

// DTO untuk membaca JSON dari Flutter
type CreateReportRequest struct {
	VillageID      string               `json:"village_id" validate:"required"`
	RT             string               `json:"rt" validate:"required"`
	RW             string               `json:"rw" validate:"required"`
	FamilyHeadName string               `json:"family_head_name"`
	Latitude       float64              `json:"latitude" validate:"required"`
	Longitude      float64              `json:"longitude" validate:"required"`
	LarvaeStatus   int                  `json:"larvae_status" validate:"oneof=0 1"`
	Containers     []ContainerDetailReq `json:"containers" validate:"required,dive"`
}

type ContainerDetailReq struct {
	ContainerType  string `json:"container_type" validate:"required"`
	InspectedCount int    `json:"inspected_count"`
	PositiveCount  int    `json:"positive_count"`
}

// 1. Fungsi Create (Untuk Form Laporan Kader)
func (h *InspectionReportHandler) Create(c *fiber.Ctx) error {
	req := new(CreateReportRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body", "details": err.Error()})
	}

	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	report := &domain.InspectionReport{
		UserID:         userID,
		VillageID:      req.VillageID,
		RT:             req.RT,
		RW:             req.RW,
		FamilyHeadName: req.FamilyHeadName,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		LarvaeStatus:   req.LarvaeStatus,
	}

	for _, cReq := range req.Containers {
		report.ContainerDetails = append(report.ContainerDetails, domain.ContainerDetail{
			ContainerType:  cReq.ContainerType,
			InspectedCount: cReq.InspectedCount,
			PositiveCount:  cReq.PositiveCount,
		})
	}

	if err := h.service.CreateReport(c.Context(), report); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create report", "details": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Report successfully created",
		"data":    report,
	})
}

// 2. Fungsi GetHistory (Untuk Halaman Riwayat Kader)
func (h *InspectionReportHandler) GetHistory(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	reports, err := h.service.GetCadreHistory(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": reports})
}

// 3. Fungsi GetPending (Untuk Halaman Validasi Petugas)
func (h *InspectionReportHandler) GetPending(c *fiber.Ctx) error {
	reports, err := h.service.GetPendingReports(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": reports})
}

// 4. Fungsi ValidateReport (Untuk Tombol Terima/Tolak Petugas)
func (h *InspectionReportHandler) ValidateReport(c *fiber.Ctx) error {
	reportID := c.Params("id")

	var req struct {
		Status string `json:"status" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := h.service.ValidateReport(c.Context(), reportID, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update validation status"})
	}
	return c.JSON(fiber.Map{"message": "Report validated successfully"})
}

// 5. Fungsi GetMapData (Untuk Halaman Peta IDW)
func (h *InspectionReportHandler) GetMapData(c *fiber.Ctx) error {
	reports, err := h.service.GetMapData(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": reports})
}
