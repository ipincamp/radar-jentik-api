package handlers

import (
	"fmt"
	"math"
	"time"

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
	PhotoURL       string               `json:"photo_url"`
	InspectedAt    string               `json:"inspected_at"`
	Containers     []ContainerDetailReq `json:"containers" validate:"required,dive"`
	// LarvaeStatus   bool                 `json:"larvae_status"`
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

type BulkValidateRequest struct {
	ReportIDs       []string `json:"report_ids" validate:"required,dive,required"`
	Status          string   `json:"status" validate:"required,oneof=accept reject"`
	RejectionReason *string  `json:"rejection_reason,omitempty"`
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

	// statusJentik := 0
	// if req.LarvaeStatus {
	// 	statusJentik = 1
	// }

	// Mapping Request ke Domain Entity
	report := &domain.InspectionReport{
		UserID:         userID,
		VillageID:      req.VillageID,
		RT:             req.RT,
		RW:             req.RW,
		FamilyHeadName: req.FamilyHeadName,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		// LarvaeStatus:   statusJentik,
		PhotoURL: req.PhotoURL,
	}

	if req.InspectedAt != "" {
		// Parse string ISO 8601 dari Flutter
		parsedTime, err := time.Parse(time.RFC3339, req.InspectedAt)
		if err == nil {
			report.InspectedAt = parsedTime // Gunakan waktu dari HP kader
		}
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

	// 2. Tangkap parameter pagination dan filter dari URL
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	search := c.Query("search")
	rt := c.Query("rt")
	rw := c.Query("rw")
	villageID := c.Query("village_id")
	date := c.Query("date")

	// 3. Panggil Service versi Paginated dengan filter
	reports, totalData, err := h.service.GetPaginatedHistory(c.Context(), userID, page, limit, search, rt, rw, villageID, date)
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
			VillageID:      r.VillageID,
			RT:             r.RT,
			RW:             r.RW,
			FamilyHeadName: r.FamilyHeadName,
			Latitude:       r.Latitude,
			Longitude:      r.Longitude,
			// LarvaeStatus:   0,
			PhotoURL: r.PhotoURL,
		}

		if r.InspectedAt != "" {
			// Parse string ISO 8601 dari Flutter
			parsedTime, err := time.Parse(time.RFC3339, r.InspectedAt)
			if err == nil {
				report.InspectedAt = parsedTime
			}
		}

		// if r.LarvaeStatus {
		// 	report.LarvaeStatus = 1
		// }

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

// BulkValidateReports menangani proses validasi massal (Terima/Tolak) banyak laporan sekaligus
func (h *InspectionReportHandler) BulkValidateReports(c *fiber.Ctx) error {
	var req BulkValidateRequest

	// 1. Parsing Body JSON ke Struct
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
			"error":   err.Error(),
		})
	}

	// (Opsional) 2. Validasi manual jika Anda tidak menggunakan library validator
	if len(req.ReportIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "report_ids tidak boleh kosong",
		})
	}
	if req.Status != "accept" && req.Status != "reject" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Status harus 'accept' atau 'reject'",
		})
	}

	// 3. Persiapkan data yang akan di-update
	updateData := map[string]interface{}{
		"validation_status": req.Status,
	}

	// Jika statusnya reject, masukkan alasan. Jika accept, kosongkan alasan penolakan.
	if req.Status == "reject" && req.RejectionReason != nil {
		updateData["rejection_reason"] = *req.RejectionReason
	} else if req.Status == "accept" {
		updateData["rejection_reason"] = nil // Hapus alasan penolakan jika sebelumnya ada
	}

	// 4. Eksekusi Update Massal menggunakan GORM (WHERE id IN (...))
	err := h.service.BulkValidateReports(c.Context(), req.ReportIDs, req.Status)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui status laporan",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("%d laporan berhasil diperbarui statusnya menjadi '%s'", len(req.ReportIDs), req.Status),
	})
}

// 8. Fungsi UpdateReport (Untuk Mengedit Laporan Kader)
func (h *InspectionReportHandler) UpdateReport(c *fiber.Ctx) error {
	reportID := c.Params("id")

	// Kita bisa meminjam DTO Create karena format body JSON-nya persis sama
	req := new(CreateReportRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	userID, _ := c.Locals("user_id").(string)

	report := &domain.InspectionReport{
		UserID:         userID,
		VillageID:      req.VillageID,
		RT:             req.RT,
		RW:             req.RW,
		FamilyHeadName: req.FamilyHeadName,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		PhotoURL:       req.PhotoURL,
	}

	if req.InspectedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.InspectedAt)
		if err == nil {
			report.InspectedAt = parsedTime
		}
	}

	for _, cReq := range req.Containers {
		report.ContainerDetails = append(report.ContainerDetails, domain.ContainerDetail{
			ContainerTypeID: cReq.ContainerTypeID,
			CustomName:      cReq.CustomName,
			InspectedCount:  cReq.InspectedCount,
			PositiveCount:   cReq.PositiveCount,
		})
	}

	if err := h.service.UpdateReport(c.Context(), reportID, report); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal memperbarui laporan", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, "Laporan berhasil diperbarui", nil)
}
