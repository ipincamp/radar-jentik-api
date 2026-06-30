package handlers

import (
	"fmt"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/utils"
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
	VillageID      string               `json:"village_id"`
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
	ContainerTypeID string  `json:"container_type_id" validate:"required"`
	CustomName      *string `json:"custom_name"`
	InspectedCount  int     `json:"inspected_count"`
	PositiveCount   int     `json:"positive_count"`
}

type BulkCreateReportRequest struct {
	Reports []CreateReportRequest `json:"reports" validate:"required,dive"`
}

// ---------------------------------------------------------
// 1. Fungsi Create (Untuk Form Laporan Kader)
// ---------------------------------------------------------
func (h *InspectionReportHandler) Create(c *fiber.Ctx) error {
	req := new(CreateReportRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
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
			ContainerTypeID: cReq.ContainerTypeID,
			CustomName:      cReq.CustomName,
			InspectedCount:  cReq.InspectedCount,
			PositiveCount:   cReq.PositiveCount,
		})
	}

	if err := h.service.CreateReport(c.Context(), report); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal membuat laporan", err.Error())
	}

	return utils.Success(c, fiber.StatusCreated, "Laporan berhasil dikirim", nil)
}

// ---------------------------------------------------------
// 2. Fungsi GetHistory (Untuk Halaman Riwayat Kader)
// ---------------------------------------------------------
func (h *InspectionReportHandler) GetHistory(c *fiber.Ctx) error {
	// 1. Ambil User ID dari token JWT (Kader yang sedang login)
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Sesi tidak valid, harap login ulang", "")
	}

	// 2. Tangkap parameter pagination dari URL (Default: page 1, limit 10)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// 3. Panggil Service versi Paginated
	reports, totalData, err := h.service.GetPaginatedHistory(c.Context(), userID, page, limit)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil riwayat laporan", err.Error())
	}

	// 4. Susun Metadata Pagination
	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))
	meta := utils.PaginationMeta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  totalData,
		TotalPages:  totalPages,
	}

	// 5. Kembalikan Response dengan format Paginated
	return utils.Paginated(c, fiber.StatusOK, "Riwayat laporan berhasil diambil", reports, meta)
}

// ---------------------------------------------------------
// 3. Fungsi GetPending (Untuk Halaman Validasi Petugas)
// ---------------------------------------------------------
func (h *InspectionReportHandler) GetPending(c *fiber.Ctx) error {
	// 1. Tangkap Query Params
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	// 2. Lempar ke Service
	reports, totalData, err := h.service.GetPaginatedPending(c.Context(), page, limit)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil laporan", err.Error())
	}

	// 3. Susun Metadata
	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))
	meta := utils.PaginationMeta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  totalData,
		TotalPages:  totalPages,
	}

	// 4. Return format standar
	return utils.Paginated(c, fiber.StatusOK, "Daftar antrean laporan", reports, meta)
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
		return utils.Error(c, fiber.StatusBadRequest, "Format data tidak valid", err.Error())
	}

	// 3. Panggil Service UpdateStatus (yang memuat logika bisnis penolakan)
	// Kita mengirimkan &req.RejectionReason (pointer) agar selaras dengan Service
	err := h.service.UpdateStatus(c.Context(), reportID, req.Status, &req.RejectionReason)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal memperbarui status laporan", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Status laporan berhasil diperbarui", nil)
}

// ---------------------------------------------------------
// 5. Fungsi GetMapData (Untuk Halaman Peta IDW)
// ---------------------------------------------------------
func (h *InspectionReportHandler) GetMapData(c *fiber.Ctx) error {

	// Ambil User ID dan Role dari Locals (hasil ekstraksi Middleware)
	userID, ok1 := c.Locals("user_id").(string)
	role, ok2 := c.Locals("role").(string)

	if !ok1 || !ok2 {
		return utils.Error(c, fiber.StatusUnauthorized, "Sesi tidak valid, harap login ulang", "")
	}

	// Masukkan userID dan role ke dalam pemanggilan Service
	reports, err := h.service.GetMapData(c.Context(), userID, role)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data peta spasial", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Data peta spasial berhasil diambil", reports)
}

// ---------------------------------------------------------
// 6. Fungsi ExportExcel (Untuk Tombol Export Excel)
// ---------------------------------------------------------
func (h *InspectionReportHandler) ExportExcel(c *fiber.Ctx) error {
	// 1. Ambil sesi user dari JWT
	userID, ok1 := c.Locals("user_id").(string)
	role, ok2 := c.Locals("role").(string)

	if !ok1 || !ok2 {
		return utils.Error(c, fiber.StatusUnauthorized, "Sesi tidak valid, harap login ulang", "")
	}

	// 2. Panggil Service untuk membuat Excel
	fileBytes, fileName, err := h.service.ExportToExcel(c.Context(), userID, role)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengekspor data ke Excel", err.Error())
	}

	// 3. Atur Header HTTP agar browser/Flutter tahu ini adalah file unduhan
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// 4. Kirim file biner-nya
	return c.Send(fileBytes)
}

// ---------------------------------------------------------
// 7. Fungsi CreateBulk (Untuk Sinkronisasi Massal Laporan Kader)
// ---------------------------------------------------------
func (h *InspectionReportHandler) CreateBulk(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Sesi tidak valid", "")
	}

	req := new(BulkCreateReportRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	// Jika array kosong
	if len(req.Reports) == 0 {
		return utils.Error(c, fiber.StatusBadRequest, "Data laporan kosong", nil)
	}

	var domainReports []*domain.InspectionReport

	// Mapping dari Request (DTO) ke Domain Array
	for _, r := range req.Reports {
		report := &domain.InspectionReport{
			UserID:         userID,
			VillageID:      r.VillageID, // Bisa kosong karena dihandle service
			RT:             r.RT,
			RW:             r.RW,
			FamilyHeadName: r.FamilyHeadName,
			Latitude:       r.Latitude,
			Longitude:      r.Longitude,
			LarvaeStatus:   0,
			PhotoURL:       r.PhotoURL,
		}

		if r.LarvaeStatus {
			report.LarvaeStatus = 1
		}

		for _, cReq := range r.Containers {
			report.ContainerDetails = append(report.ContainerDetails, domain.ContainerDetail{
				ContainerTypeID: cReq.ContainerTypeID,
				CustomName:      cReq.CustomName,
				InspectedCount:  cReq.InspectedCount,
				PositiveCount:   cReq.PositiveCount,
			})
		}

		domainReports = append(domainReports, report)
	}

	// Lempar ke Service
	if err := h.service.CreateBulkReport(c.Context(), domainReports); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal memproses sinkronisasi massal", err.Error())
	}

	return utils.Success(c, fiber.StatusCreated, fmt.Sprintf("%d laporan berhasil disinkronisasi", len(domainReports)), nil)
}
