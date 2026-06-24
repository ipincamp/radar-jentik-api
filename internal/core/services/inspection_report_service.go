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

// Struct pembantu untuk rekap per RT
type RecapRT struct {
	RT             string
	DawisTotal     int
	DawisAktif     int
	RumahDiperiksa int
	RumahPositif   int
	// Jenis Container [Total, Positif]
	BakMandi     [2]int
	Tempayan     [2]int
	PecahanBotol [2]int
	BarangBekas  [2]int
	Kulkas       [2]int
	TandonAir    [2]int
	VasBunga     [2]int
	PotBunga     [2]int
	LainLain     [2]int
	// Total Kontainer
	TotalContainer    int
	TotalContainerPos int
}

func NewInspectionReportService(repo ports.InspectionReportRepository) ports.InspectionReportService {
	return &inspectionReportService{
		repo: repo,
	}
}

// CreateReport bertugas memproses payload dan mengirimkannya ke repository
func (s *inspectionReportService) CreateReport(ctx context.Context, report *domain.InspectionReport) error {
	// Anda bisa menambahkan logika bisnis di sini (misal: validasi tambahan, perhitungan, dll)

	// Set default value sebelum masuk ke database
	report.ValidationStatus = "pending"
	report.InspectedAt = time.Now()

	// Memanggil repository yang sudah membungkus proses insert ke dalam Transaksi SQL
	err := s.repo.Create(ctx, report)
	if err != nil {
		return err
	}

	return nil
}

func (s *inspectionReportService) GetCadreHistory(ctx context.Context, userID string) ([]domain.InspectionReport, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *inspectionReportService) GetPendingReports(ctx context.Context) ([]domain.InspectionReport, error) {
	return s.repo.GetPending(ctx)
}

func (s *inspectionReportService) ValidateReport(ctx context.Context, id string, status string) error {
	// Memastikan status yang masuk hanya 'accept' atau 'reject'
	if status != "accept" && status != "reject" {
		return context.DeadlineExceeded // atau buat custom error
	}
	return s.repo.UpdateStatus(ctx, id, status, nil)
}

func (s *inspectionReportService) GetMapData(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error) {
	// Teruskan parameter userID dan role ke Repository
	return s.repo.GetValidReports(ctx, userID, role)
}

func (s *inspectionReportService) ExportToExcel(ctx context.Context, userID, role string) ([]byte, error) {
	// 1. Ambil data mentah langsung dari Database
	recapData, err := s.repo.GetRecapData(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	// ==========================================
	// BAGIAN 1: HEADER INFORMASI (KABUPATEN, DESA, BULAN)
	// ==========================================
	f.SetCellValue(sheet, "A1", "Kabupaten")
	f.SetCellValue(sheet, "C1", ": Banyumas") // Ubah sesuai data

	f.SetCellValue(sheet, "A2", "Kecamatan")
	f.SetCellValue(sheet, "C2", ": Cilongok")

	f.SetCellValue(sheet, "A3", "Kelurahan/Desa")
	f.SetCellValue(sheet, "C3", ": Langgongsari")

	f.SetCellValue(sheet, "A4", "Bulan/Tahun")
	f.SetCellValue(sheet, "C4", ": Mei 2026")

	f.SetCellValue(sheet, "A5", "RW")
	f.SetCellValue(sheet, "C5", ": 01")

	f.SetCellValue(sheet, "A6", "Jumlah Rumah")
	f.SetCellValue(sheet, "C6", ": 95")

	// ==========================================
	// BAGIAN 2: MEMBUAT HEADER TABEL (MERGED CELLS)
	// ==========================================
	// Baris 8 adalah Judul Utama, Baris 9 adalah Sub-Judul (Jml / Pos)

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

	// Header Container Besar (Merge F sampai W)
	f.SetCellValue(sheet, "F8", "Container")
	f.MergeCell(sheet, "F8", "W8")

	// Sub-Header Container
	containers := []string{
		"Bak Mandi", "Tempayan", "Pc. Botol", "Brg Bekas",
		"Kulkas", "Tandon Air", "Vas Bunga", "Pot Bunga", "Lain-lain",
	}

	colAscii := 70 // Kode ASCII untuk 'F'
	for _, name := range containers {
		colName1 := string(rune(colAscii))     // Contoh: F
		colName2 := string(rune(colAscii + 1)) // Contoh: G

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

	// Styling Header supaya Bold, Tengah, dan Text Wrap
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A8", "Z9", headerStyle)

	// ==========================================
	// BAGIAN 3: MENGISI DATA & RUMUS BAWAH
	// ==========================================
	startRow := 10
	var totalRumahDiperiksa, totalRumahPositif float64
	var totalKontainerSeluruh, totalKontainerPositif float64

	for i, r := range recapData {
		row := startRow + i

		// Hitung CI per RT
		ciRT := 0.0
		if r.TotalContainer > 0 {
			ciRT = (float64(r.TotalContainerPos) / float64(r.TotalContainer)) * 100
		}

		// Masukkan ke kolom.
		// Catatan: Dawis tidak ada di skema Database, kita set "0" secara default.
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.RT)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), 0) // DawisTotal
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), 0) // DawisAktif
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.RumahDiperiksa)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.RumahPositif)

		// Rincian Container yang di-map dari hasil query SQL
		cols := []int{
			r.BakMandiTotal, r.BakMandiPos, r.TempayanTotal, r.TempayanPos,
			r.PecahanBotolTotal, r.PecahanBotolPos, r.BarangBekasTotal, r.BarangBekasPos,
			r.KulkasTotal, r.KulkasPos, r.TandonAirTotal, r.TandonAirPos,
			r.VasBungaTotal, r.VasBungaPos, r.PotBungaTotal, r.PotBungaPos,
			r.LainLainTotal, r.LainLainPos,
		}

		colAsciiData := 70 // Mulai dari kolom 'F'
		for _, val := range cols {
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", string(rune(colAsciiData)), row), val)
			colAsciiData++
		}

		f.SetCellValue(sheet, fmt.Sprintf("X%d", row), r.TotalContainer)
		f.SetCellValue(sheet, fmt.Sprintf("Y%d", row), r.TotalContainerPos)
		f.SetCellValue(sheet, fmt.Sprintf("Z%d", row), fmt.Sprintf("%.2f %%", ciRT))

		// Akumulasi untuk Footer ABJ & CI Keseluruhan
		totalRumahDiperiksa += float64(r.RumahDiperiksa)
		totalRumahPositif += float64(r.RumahPositif)
		totalKontainerSeluruh += float64(r.TotalContainer)
		totalKontainerPositif += float64(r.TotalContainerPos)
	}

	// ==========================================
	// BAGIAN 4: FOOTER (REKAP ABJ & CI)
	// ==========================================
	footerRow := startRow + len(recapData) + 2

	// Hitung ABJ = ((Diperiksa - Positif) / Diperiksa) * 100%
	abj := 0.0
	if totalRumahDiperiksa > 0 {
		abj = ((totalRumahDiperiksa - totalRumahPositif) / totalRumahDiperiksa) * 100
	}

	// Hitung Total CI = (Total Positif / Total Container) * 100%
	ci := 0.0
	if totalKontainerSeluruh > 0 {
		ci = (totalKontainerPositif / totalKontainerSeluruh) * 100
	}

	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow), "REKAPITULASI KESELURUHAN")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("A%d", footerRow), headerStyle) // Bold

	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow+1), "Angka Bebas Jentik (ABJ)")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", footerRow+1), fmt.Sprintf(": %.2f %%", abj))

	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow+2), "Container Index (CI)")
	f.SetCellValue(sheet, fmt.Sprintf("C%d", footerRow+2), fmt.Sprintf(": %.2f %%", ci))

	// 5. Tulis file Excel ke dalam memory buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *inspectionReportService) UpdateStatus(ctx context.Context, reportID string, status string, rejectionReason *string) error {
	var reasonPtr *string

	if status == "reject" {
		if rejectionReason == nil || *rejectionReason == "" {
			return errors.New("alasan penolakan wajib diisi")
		}
		reasonPtr = rejectionReason
	} else if status == "accept" {
		// Jika diterima, pastikan alasannya NULL
		reasonPtr = nil
	} else {
		return errors.New("status tidak valid")
	}

	return s.repo.UpdateStatus(ctx, reportID, status, reasonPtr)
}
