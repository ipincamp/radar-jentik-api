package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type IDWHandler struct {
	idwService    ports.IDWService
	reportService ports.InspectionReportService
}

// Tambahkan reportService ke dalam constructor
func NewIDWHandler(idwService ports.IDWService, reportService ports.InspectionReportService) *IDWHandler {
	return &IDWHandler{
		idwService:    idwService,
		reportService: reportService,
	}
}

func (h *IDWHandler) Calculate(c *fiber.Ctx) error {
	var req domain.IDWRequest

	// 1. Parsing Bounding Box (Batas Area) dari Flutter/Postman
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format payload tidak valid",
		})
	}

	// 2. Ambil Sesi User (Role & ID) dari Middleware JWT
	userID, ok1 := c.Locals("user_id").(string)
	role, ok2 := c.Locals("role").(string)

	if !ok1 || !ok2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Sesi tidak valid"})
	}

	// 3. Ambil Data Inspeksi Riil dari Database (Hanya yang 'accept')
	reports, err := h.reportService.GetMapData(c.Context(), userID, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data laporan inspeksi",
		})
	}

	// 4. Transformasi Data Laporan menjadi Titik Sampel IDW
	var samples []domain.SamplePoint
	for _, r := range reports {
		// Logika Konversi Kerawanan:
		// Jika LarvaeStatus == 1 (Ada Jentik) -> Nilai Kerawanan = 100 (Merah)
		// Jika LarvaeStatus == 0 (Bebas Jentik) -> Nilai Kerawanan = 0 (Hijau)
		var value float64 = 0.0
		if r.LarvaeStatus == 1 {
			value = 100.0
		}

		samples = append(samples, domain.SamplePoint{
			Lat:   r.Latitude,
			Lon:   r.Longitude,
			Value: value,
		})
	}

	// Cek apakah ada data sampel di database
	if len(samples) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Belum ada data laporan inspeksi tervalidasi untuk area ini",
		})
	}

	// Timpa req.Samples bawaan Postman dengan data riil dari database
	req.Samples = samples

	// Default parameter jika tidak dikirim dari frontend
	if req.Power <= 0 {
		req.Power = 2.0
	}
	// Default resolution
	if req.Resolution <= 0 {
		req.Resolution = 0.001
	}

	// 5. Kalkulasi IDW menggunakan Service
	gridResult, err := h.idwService.CalculateIDWGrid(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghitung estimasi IDW",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Berhasil menghitung IDW berdasarkan data riil",
		"data":    gridResult,
	})
}
