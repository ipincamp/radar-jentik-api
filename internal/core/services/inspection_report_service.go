package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/xuri/excelize/v2"
)

type inspectionReportService struct {
	repo ports.InspectionReportRepository
}

func NewInspectionReportService(repo ports.InspectionReportRepository) ports.InspectionReportService {
	return &inspectionReportService{
		repo: repo,
	}
}

// CreateReport bertugas memproses payload dan memvalidasi aturan bisnis
func (s *inspectionReportService) CreateReport(ctx context.Context, report *domain.InspectionReport) error {
	// 1. VALIDASI BISNIS: Wajib melampirkan URL Foto (Permintaan Petugas)
	if report.PhotoURL == "" {
		return errors.New("foto bukti inspeksi lapangan wajib disertakan")
	}

	// 2. VALIDASI BISNIS: Minimal harus ada 1 wadah yang dilaporkan
	if len(report.ContainerDetails) == 0 {
		return errors.New("minimal harus melaporkan satu jenis wadah")
	}

	// 3. Set default value sebelum masuk ke database
	report.ValidationStatus = "pending"
	report.InspectedAt = time.Now()

	// 4. Kirim ke repository untuk disimpan
	return s.repo.Create(ctx, report)
}

func (s *inspectionReportService) GetCadreHistory(ctx context.Context, userID string) ([]domain.InspectionReport, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *inspectionReportService) GetPendingReports(ctx context.Context) ([]domain.InspectionReport, error) {
	return s.repo.GetPending(ctx)
}

func (s *inspectionReportService) ValidateReport(ctx context.Context, id string, status string) error {
	if status != "accept" && status != "reject" {
		return errors.New("status validasi tidak dikenali")
	}
	return s.repo.UpdateStatus(ctx, id, status, nil)
}

func (s *inspectionReportService) UpdateStatus(ctx context.Context, reportID string, status string, rejectionReason *string) error {
	var reasonPtr *string

	if status == "reject" {
		if rejectionReason == nil || *rejectionReason == "" {
			return errors.New("alasan penolakan wajib diisi jika menolak laporan")
		}
		reasonPtr = rejectionReason
	} else if status == "accept" {
		reasonPtr = nil // Bersihkan alasan jika diterima
	} else {
		return errors.New("status tidak valid")
	}

	return s.repo.UpdateStatus(ctx, reportID, status, reasonPtr)
}

func (s *inspectionReportService) GetMapData(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error) {
	return s.repo.GetValidReports(ctx, userID, role)
}

func (s *inspectionReportService) ExportToExcel(ctx context.Context, userID, role string) ([]byte, error) {
	// Peringatan Backend: Karena sekarang kita menggunakan container_type_id (Normalisasi),
	// Nanti di layer Repository (GetRecapData) kita wajib melakukan SQL JOIN ke tabel container_types
	// agar data Rekap tetap akurat. Logika Excelize di bawah ini tidak perlu diubah.
	recapData, err := s.repo.GetRecapData(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	// --- SETUP HEADER EXCEL ---
	f.SetCellValue(sheet, "A1", "Kabupaten")
	f.SetCellValue(sheet, "C1", ": Banyumas")
	f.SetCellValue(sheet, "A2", "Kecamatan")
	f.SetCellValue(sheet, "C2", ": Cilongok")
	f.SetCellValue(sheet, "A3", "Kelurahan/Desa")
	f.SetCellValue(sheet, "C3", ": Langgongsari")
	f.SetCellValue(sheet, "A4", "Bulan/Tahun")
	f.SetCellValue(sheet, "C4", ": Laporan Terbaru")
	f.SetCellValue(sheet, "A5", "RW")
	f.SetCellValue(sheet, "C5", ": 01")

	// HEADER TABEL (Merged Cells)
	f.SetCellValue(sheet, "A8", "RT")
	f.MergeCell(sheet, "A8", "A9")
	f.SetCellValue(sheet, "B8", "Jumlah Dawis")
	f.MergeCell(sheet, "B8", "C8")
	f.SetCellValue(sheet, "B9", "Seluruh")
	f.SetCellValue(sheet, "C9", "Aktif")
	f.SetCellValue(sheet, "D8", "Jumlah Rumah\nDiperiksa")
	f.MergeCell(sheet, "D8", "D9")
	f.SetCellValue(sheet, "E8", "Jumlah Rumah\nPositip")
	f.MergeCell(sheet, "E8", "E9")

	f.SetCellValue(sheet, "F8", "Container")
	f.MergeCell(sheet, "F8", "W8")

	containers := []string{
		"Bak Mandi", "Tempayan", "Pc. Botol", "Brg Bekas",
		"Kulkas", "Tandon Air", "Vas Bunga", "Pot Bunga", "Lain-lain",
	}

	colAscii := 70 // Kode ASCII untuk 'F'
	for _, name := range containers {
		colName1 := string(rune(colAscii))
		colName2 := string(rune(colAscii + 1))
		f.SetCellValue(sheet, colName1+"9", name+"\n(Jml)")
		f.SetCellValue(sheet, colName2+"9", name+"\n(Pos)")
		colAscii += 2
	}

	f.SetCellValue(sheet, "X8", "Jumlah Kontainer")
	f.MergeCell(sheet, "X8", "Y8")
	f.SetCellValue(sheet, "X9", "Jml")
	f.SetCellValue(sheet, "Y9", "Pos")
	f.SetCellValue(sheet, "Z8", "Container\nIndex (CI)")
	f.MergeCell(sheet, "Z8", "Z9")

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A8", "Z9", headerStyle)

	// --- MENGISI DATA ---
	startRow := 10
	var totalRumahDiperiksa, totalRumahPositif float64
	var totalKontainerSeluruh, totalKontainerPositif float64

	for i, r := range recapData {
		row := startRow + i

		ciRT := 0.0
		if r.TotalContainer > 0 {
			ciRT = (float64(r.TotalContainerPos) / float64(r.TotalContainer)) * 100
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.RT)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), 0)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), 0)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.RumahDiperiksa)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.RumahPositif)

		cols := []int{
			r.BakMandiTotal, r.BakMandiPos, r.TempayanTotal, r.TempayanPos,
			r.PecahanBotolTotal, r.PecahanBotolPos, r.BarangBekasTotal, r.BarangBekasPos,
			r.KulkasTotal, r.KulkasPos, r.TandonAirTotal, r.TandonAirPos,
			r.VasBungaTotal, r.VasBungaPos, r.PotBungaTotal, r.PotBungaPos,
			r.LainLainTotal, r.LainLainPos,
		}

		colAsciiData := 70
		for _, val := range cols {
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", string(rune(colAsciiData)), row), val)
			colAsciiData++
		}

		f.SetCellValue(sheet, fmt.Sprintf("X%d", row), r.TotalContainer)
		f.SetCellValue(sheet, fmt.Sprintf("Y%d", row), r.TotalContainerPos)
		f.SetCellValue(sheet, fmt.Sprintf("Z%d", row), fmt.Sprintf("%.2f %%", ciRT))

		totalRumahDiperiksa += float64(r.RumahDiperiksa)
		totalRumahPositif += float64(r.RumahPositif)
		totalKontainerSeluruh += float64(r.TotalContainer)
		totalKontainerPositif += float64(r.TotalContainerPos)
	}

	// --- FOOTER REKAP ---
	footerRow := startRow + len(recapData) + 2
	abj, ci := 0.0, 0.0
	if totalRumahDiperiksa > 0 {
		abj = ((totalRumahDiperiksa - totalRumahPositif) / totalRumahDiperiksa) * 100
	}
	if totalKontainerSeluruh > 0 {
		ci = (totalKontainerPositif / totalKontainerSeluruh) * 100
	}

	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow), "REKAPITULASI KESELURUHAN")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("A%d", footerRow), headerStyle)
	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow+1), "Angka Bebas Jentik (ABJ)")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", footerRow+1), fmt.Sprintf(": %.2f %%", abj))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow+2), "Container Index (CI)")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", footerRow+2), fmt.Sprintf(": %.2f %%", ci))

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
