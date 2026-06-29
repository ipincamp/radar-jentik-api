package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type InspectionReportHandler struct {
	service ports.InspectionReportService
}

func NewInspectionReportHandler(service ports.InspectionReportService) *InspectionReportHandler {
	return &InspectionReportHandler{service: service}
}

// ---------------------------------------------------------
// DTO (Data Transfer Object) - Penyesuaian Request Body
// ---------------------------------------------------------

type CreateReportRequest struct {
	VillageID      string               `json:"village_id" validate:"required"`
	RT             string               `json:"rt" validate:"required"`
	RW             string               `json:"rw" validate:"required"`
	FamilyHeadName string               `json:"family_head_name"`
	Latitude       float64              `json:"latitude" validate:"required"`
	Longitude      float64              `json:"longitude" validate:"required"`
	LarvaeStatus   bool                 `json:"larvae_status"`
	PhotoURL       string               `json:"photo_url" validate:"required"`
	Containers     []ContainerDetailReq `json:"containers" validate:"required,dive"`
}

type ContainerDetailReq struct {
	ContainerTypeID string `json:"container_type_id" validate:"required"`
	InspectedCount  int    `json:"inspected_count"`
	PositiveCount   int    `json:"positive_count"`
}

// ---------------------------------------------------------
// 1. Fungsi Create (Untuk Form Laporan Kader)
// ---------------------------------------------------------
func (h *InspectionReportHandler) Create(c *fiber.Ctx) error {
	req := new(CreateReportRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	userID, _ := c.Locals("user_id").(string)

	statusJentik := 0
	if req.LarvaeStatus {
		statusJentik = 1
	}

	// Mapping Request ke Domain Entity
	report := &domain.InspectionReport{
		UserID:         userID,
		VillageID:      req.VillageID,
		RT:             req.RT,
		RW:             req.RW,
		FamilyHeadName: req.FamilyHeadName,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		LarvaeStatus:   statusJentik,
		PhotoURL:       req.PhotoURL,
	}

	// Mapping Container Details
	for _, cReq := range req.Containers {
		report.ContainerDetails = append(report.ContainerDetails, domain.ContainerDetail{
			ContainerTypeID: cReq.ContainerTypeID, // ✨ Map ID Wadah
			InspectedCount:  cReq.InspectedCount,
			PositiveCount:   cReq.PositiveCount,
		})
	}

	if err := h.service.CreateReport(c.Context(), report); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Laporan berhasil dikirim"})
}

// ---------------------------------------------------------
// 2. Fungsi GetHistory (Untuk Halaman Riwayat Kader)
// ---------------------------------------------------------
func (h *InspectionReportHandler) GetHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID tidak valid atau tidak ditemukan"})
	}

	reports, err := h.service.GetCadreHistory(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil riwayat laporan"})
	}

	return c.JSON(fiber.Map{"data": reports})
}

// ---------------------------------------------------------
// 3. Fungsi GetPending (Untuk Halaman Validasi Petugas)
// ---------------------------------------------------------
func (h *InspectionReportHandler) GetPending(c *fiber.Ctx) error {
	reports, err := h.service.GetPendingReports(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil daftar laporan pending"})
	}

	return c.JSON(fiber.Map{"data": reports})
}

// ---------------------------------------------------------
// 4. Fungsi ValidateReport (Untuk Tombol Terima/Tolak Petugas)
// ---------------------------------------------------------
func (h *InspectionReportHandler) ValidateReport(c *fiber.Ctx) error {
	// 1. Ambil ID Laporan dari URL (misal: /api/v1/reports/:id/validate)
	reportID := c.Params("id")

	// 2. Parsing Body JSON
	var req domain.ValidateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format data tidak valid",
		})
	}

	// 3. Panggil Service
	// err := h.reportService.UpdateStatus(c.Context(), reportID, req.Status, req.RejectionReason)
	err := h.service.ValidateReport(c.Context(), reportID, req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Status laporan berhasil diperbarui",
	})
}

// ---------------------------------------------------------
// 5. Fungsi GetMapData (Untuk Halaman Peta IDW)
// ---------------------------------------------------------
func (h *InspectionReportHandler) GetMapData(c *fiber.Ctx) error {

	// Ambil User ID dan Role dari Locals (hasil ekstraksi Middleware)
	userID, ok1 := c.Locals("user_id").(string)
	role, ok2 := c.Locals("role").(string)

	if !ok1 || !ok2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Sesi tidak valid, harap login ulang"})
	}

	// Masukkan userID dan role ke dalam pemanggilan Service
	reports, err := h.service.GetMapData(c.Context(), userID, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data peta spasial"})
	}

	return c.JSON(fiber.Map{"data": reports})
}

func (h *InspectionReportHandler) ExportExcel(c *fiber.Ctx) error {
	// 1. Ambil sesi user dari JWT
	userID, ok1 := c.Locals("user_id").(string)
	role, ok2 := c.Locals("role").(string)

	if !ok1 || !ok2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Sesi tidak valid"})
	}

	// 2. Panggil Service untuk membuat Excel
	fileBytes, err := h.service.ExportToExcel(c.Context(), userID, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal men-generate file Excel",
		})
	}

	// 3. Atur Header HTTP agar browser/Flutter tahu ini adalah file unduhan
	// Content-Type khusus format .xlsx
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// Nama file saat diunduh
	c.Set("Content-Disposition", "attachment; filename=Laporan_Radar_Jentik.xlsx")

	// 4. Kirim file biner-nya
	return c.Send(fileBytes)
}
