package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/xuri/excelize/v2"
)

type inspectionReportService struct {
	repo        ports.InspectionReportRepository
	villageRepo ports.VillageRepository
}

func NewInspectionReportService(repo ports.InspectionReportRepository, villageRepo ports.VillageRepository) ports.InspectionReportService {
	return &inspectionReportService{
		repo:        repo,
		villageRepo: villageRepo,
	}
}

// CreateReport bertugas memproses payload dan memvalidasi aturan bisnis
func (s *inspectionReportService) CreateReport(ctx context.Context, report *domain.InspectionReport) error {
	// Validasi input
	if report.PhotoURL == "" {
		return errors.New("foto bukti inspeksi lapangan wajib disertakan")
	}
	// Validasi minimal harus melaporkan satu jenis wadah
	if len(report.ContainerDetails) == 0 {
		return errors.New("minimal harus melaporkan satu jenis wadah")
	}

	if report.VillageID == "" {
		// Jika kader/frontend TIDAK mengirimkan ID Desa, kita auto-detect dari koordinat
		village, err := s.villageRepo.GetByCoordinate(ctx, report.Latitude, report.Longitude)
		if err != nil {
			return errors.New("lokasi anda berada di luar wilayah, silakan pilih desa secara manual")
		}
		// Isi secara otomatis
		report.VillageID = village.ID
	}

	// Set default value
	report.ValidationStatus = "pending"
	report.InspectedAt = time.Now()

	// Simpan ke database
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

func (s *inspectionReportService) ExportToExcel(ctx context.Context, userID, role string) ([]byte, string, error) {
	// 1. Ambil Data Detail dari Repository
	reports, err := s.repo.GetExportData(ctx, userID, role)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer f.Close()

	timeStr := time.Now().Format("02Jan2006_1504")

	if len(reports) == 0 {
		var buf bytes.Buffer
		f.Write(&buf)
		// Jika kosong, beri nama file kosong
		return buf.Bytes(), fmt.Sprintf("RadarJentik_%s_Kosong.xlsx", timeStr), nil
	}

	defaultSheet := f.GetSheetName(f.GetActiveSheetIndex())

	// 2. Kelompokkan Data HANYA berdasarkan [Desa]
	groupedReports := make(map[string][]domain.InspectionReport)
	var orderedVillages []string

	for _, r := range reports {
		villageName := r.Village.Name
		if _, exists := groupedReports[villageName]; !exists {
			orderedVillages = append(orderedVillages, villageName)
		}
		groupedReports[villageName] = append(groupedReports[villageName], r)
	}

	containers := []string{
		"Bak Kamar Mandi", "Tempayan", "Pecahan Botol/Air Kemasan", "Barang Bekas",
		"Kulkas/Dispenser", "Tandon Air", "Vas Bunga", "Pot Bunga", "Lain-lain",
	}

	// Array statis untuk memetakan kolom melebihi huruf 'Z' (Mencegah Bug ASCII)
	excelCols := []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P",
		"Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// 3. Buat Sheet Baru untuk Setiap Desa
	for _, villageName := range orderedVillages {
		sheetName := villageName
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		f.NewSheet(sheetName)

		// --- HEADER AKTUAL DESA ---
		f.SetCellValue(sheetName, "A1", "Kabupaten")
		f.SetCellValue(sheetName, "C1", ": Banyumas")
		f.SetCellValue(sheetName, "A2", "Kecamatan")
		f.SetCellValue(sheetName, "C2", ": Cilongok")
		f.SetCellValue(sheetName, "A3", "Kelurahan/Desa")
		f.SetCellValue(sheetName, "C3", ": "+villageName)
		f.SetCellValue(sheetName, "A4", "Tanggal Unduh")
		f.SetCellValue(sheetName, "C4", ": "+time.Now().Format("02 Jan 2006, 15:04 WIB"))

		// --- TABEL HEADER ---
		f.SetCellValue(sheetName, "A7", "No")
		f.MergeCell(sheetName, "A7", "A8")

		f.SetCellValue(sheetName, "B7", "Nama Kepala Keluarga")
		f.MergeCell(sheetName, "B7", "B8")

		f.SetCellValue(sheetName, "C7", "RW")
		f.MergeCell(sheetName, "C7", "C8")

		f.SetCellValue(sheetName, "D7", "RT")
		f.MergeCell(sheetName, "D7", "D8")

		f.SetCellValue(sheetName, "E7", "Waktu Diterima")
		f.MergeCell(sheetName, "E7", "E8")

		f.SetCellValue(sheetName, "F7", "Latitude")
		f.MergeCell(sheetName, "F7", "F8")

		f.SetCellValue(sheetName, "G7", "Longitude")
		f.MergeCell(sheetName, "G7", "G8")

		// Kolom Wadah (Huruf H sampai Y)
		f.SetCellValue(sheetName, "H7", "Container (Wadah)")
		f.MergeCell(sheetName, "H7", "Y7")

		colIdx := 7 // Index 7 merujuk ke huruf 'H' di excelCols
		for _, name := range containers {
			colName1 := excelCols[colIdx]
			colName2 := excelCols[colIdx+1]
			f.SetCellValue(sheetName, colName1+"8", name+"\n(Jml)")
			f.SetCellValue(sheetName, colName2+"8", name+"\n(Pos)")
			colIdx += 2
		}

		// Total Wadah di Z dan AA
		f.SetCellValue(sheetName, "Z7", "Total Wadah")
		f.MergeCell(sheetName, "Z7", "AA7")
		f.SetCellValue(sheetName, "Z8", "Jml")
		f.SetCellValue(sheetName, "AA8", "Pos")

		// Status Jentik di AB
		f.SetCellValue(sheetName, "AB7", "Status Jentik")
		f.MergeCell(sheetName, "AB7", "AB8")

		// Terapkan Styling Kotak/Garis Header
		f.SetCellStyle(sheetName, "A7", "AB8", headerStyle)

		// --- ISI TABEL (DATA DETAIL PER RUMAH) ---
		startRow := 9
		totalRumahDiperiksa := len(groupedReports[villageName])
		totalRumahPositif := 0
		totalKontainerSeluruh := 0
		totalKontainerPositif := 0

		for i, r := range groupedReports[villageName] {
			row := startRow + i
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), r.FamilyHeadName)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), r.RW)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), r.RT)

			// Mengisi Waktu Diterima dengan format mudah dibaca (Misal: 30 Jun 2026, 14:05 WIB)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), r.CreatedAt.Format("02 Jan 2006, 15:04 WIB"))

			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), r.Latitude)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), r.Longitude)

			if r.LarvaeStatus == 1 {
				totalRumahPositif++
				f.SetCellValue(sheetName, fmt.Sprintf("AB%d", row), "Positif")
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("AB%d", row), "Negatif")
			}

			type ContainerAgg struct {
				Inspected int
				Positive  int
			}

			cMap := make(map[string]ContainerAgg)

			for _, cd := range r.ContainerDetails {
				if cd.ContainerType != nil {
					typeName := cd.ContainerType.Name

					// Ambil data agregasi saat ini (jika belum ada, nilai default-nya 0)
					currentAgg := cMap[typeName]

					// Tambahkan jumlah dari baris database saat ini
					currentAgg.Inspected += cd.InspectedCount
					currentAgg.Positive += cd.PositiveCount

					// Simpan kembali ke map
					cMap[typeName] = currentAgg
				}
			}

			rowContainerTotal := 0
			rowContainerPos := 0
			colDataIdx := 7 // Index 7 = Huruf 'H'

			for _, cName := range containers {
				agg, exists := cMap[cName] // Mengambil struct agregasi
				colName1 := excelCols[colDataIdx]
				colName2 := excelCols[colDataIdx+1]

				if exists {
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName1, row), agg.Inspected)
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName2, row), agg.Positive)
					rowContainerTotal += agg.Inspected
					rowContainerPos += agg.Positive
				} else {
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName1, row), 0)
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName2, row), 0)
				}
				colDataIdx += 2
			}

			f.SetCellValue(sheetName, fmt.Sprintf("Z%d", row), rowContainerTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("AA%d", row), rowContainerPos)

			totalKontainerSeluruh += rowContainerTotal
			totalKontainerPositif += rowContainerPos
		}

		// --- FOOTER (REKAP KESELURUHAN DESA TSB) ---
		footerRow := startRow + len(groupedReports[villageName]) + 2
		abj, ci := 0.0, 0.0
		if totalRumahDiperiksa > 0 {
			abj = (float64(totalRumahDiperiksa-totalRumahPositif) / float64(totalRumahDiperiksa)) * 100
		}
		if totalKontainerSeluruh > 0 {
			ci = (float64(totalKontainerPositif) / float64(totalKontainerSeluruh)) * 100
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow), "REKAPITULASI DESA "+strings.ToUpper(villageName))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("A%d", footerRow), headerStyle)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow+1), "Total Rumah Diperiksa")
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", footerRow+1), fmt.Sprintf(": %d Rumah", totalRumahDiperiksa))

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow+2), "Total Rumah Positif Jentik")
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", footerRow+2), fmt.Sprintf(": %d Rumah", totalRumahPositif))

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow+3), "Angka Bebas Jentik (ABJ)")
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", footerRow+3), fmt.Sprintf(": %.2f %%", abj))

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow+4), "Container Index (CI)")
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", footerRow+4), fmt.Sprintf(": %.2f %%", ci))
	}

	f.DeleteSheet(defaultSheet)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}
	villageNameForFile := "Semua_Desa" // Default jika petugas mendownload seluruh desa sekaligus
	if len(orderedVillages) == 1 {
		// Jika hanya ada 1 desa (misal diunduh oleh kader), gunakan nama desa tersebut.
		// Spasi diganti underscore agar nama file rapi (misal: "Batu Anten" -> "Batu_Anten")
		villageNameForFile = "Desa_" + strings.ReplaceAll(orderedVillages[0], " ", "_")
	}

	fileName := fmt.Sprintf("RadarJentik_%s_%s.xlsx", timeStr, villageNameForFile)

	return buf.Bytes(), fileName, nil
}

// History
func (s *inspectionReportService) GetPaginatedHistory(ctx context.Context, userID string, page, limit int) ([]domain.InspectionReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetPaginatedHistory(ctx, userID, page, limit)
}

// Pending
func (s *inspectionReportService) GetPaginatedPending(ctx context.Context, page, limit int) ([]domain.InspectionReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetPaginatedPending(ctx, page, limit)
}
