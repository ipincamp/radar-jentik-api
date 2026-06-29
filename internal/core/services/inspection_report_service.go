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
	// 1. Ambil Data Detail
	reports, err := s.repo.GetExportData(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	// Jika kosong, kembalikan file kosong
	if len(reports) == 0 {
		var buf bytes.Buffer
		f.Write(&buf)
		return buf.Bytes(), nil
	}

	defaultSheet := f.GetSheetName(f.GetActiveSheetIndex())

	// 2. Kelompokkan Data berdasarkan [Desa + RW]
	type GroupKey struct {
		Village string
		RW      string
	}
	groupedReports := make(map[GroupKey][]domain.InspectionReport)
	var orderedKeys []GroupKey // Menjaga urutan grup

	for _, r := range reports {
		key := GroupKey{Village: r.Village.Name, RW: r.RW}
		if _, exists := groupedReports[key]; !exists {
			orderedKeys = append(orderedKeys, key)
		}
		groupedReports[key] = append(groupedReports[key], r)
	}

	containers := []string{
		"Bak Kamar Mandi", "Tempayan", "Pecahan Botol/Air Kemasan", "Barang Bekas",
		"Kulkas/Dispenser", "Tandon Air", "Vas Bunga", "Pot Bunga", "Lain-lain",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// 3. Buat Sheet Baru untuk Setiap Kelompok (Setiap RW)
	for _, key := range orderedKeys {
		// Nama sheet Excel maksimal 31 karakter
		sheetName := fmt.Sprintf("%s - RW %s", key.Village, key.RW)
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		f.NewSheet(sheetName)

		// --- HEADER ---
		f.SetCellValue(sheetName, "A1", "Kabupaten")
		f.SetCellValue(sheetName, "C1", ": Banyumas")
		f.SetCellValue(sheetName, "A2", "Kecamatan")
		f.SetCellValue(sheetName, "C2", ": Cilongok")
		f.SetCellValue(sheetName, "A3", "Kelurahan/Desa")
		f.SetCellValue(sheetName, "C3", ": "+key.Village)
		f.SetCellValue(sheetName, "A4", "RW")
		f.SetCellValue(sheetName, "C4", ": "+key.RW)
		f.SetCellValue(sheetName, "A5", "Tanggal Unduh")
		f.SetCellValue(sheetName, "C5", ": "+time.Now().Format("02 Jan 2006"))

		// --- TABEL HEADER (1 BARIS = 1 RUMAH) ---
		f.SetCellValue(sheetName, "A8", "No")
		f.MergeCell(sheetName, "A8", "A9")

		f.SetCellValue(sheetName, "B8", "Nama Kepala Keluarga")
		f.MergeCell(sheetName, "B8", "B9")

		f.SetCellValue(sheetName, "C8", "RT")
		f.MergeCell(sheetName, "C8", "C9")

		f.SetCellValue(sheetName, "D8", "Latitude")
		f.MergeCell(sheetName, "D8", "D9")

		f.SetCellValue(sheetName, "E8", "Longitude")
		f.MergeCell(sheetName, "E8", "E9")

		f.SetCellValue(sheetName, "F8", "Container (Wadah)")
		f.MergeCell(sheetName, "F8", "W8")

		colAscii := 70 // Kode ASCII huruf 'F'
		for _, name := range containers {
			colName1 := string(rune(colAscii))
			colName2 := string(rune(colAscii + 1))
			f.SetCellValue(sheetName, colName1+"9", name+"\n(Jml)")
			f.SetCellValue(sheetName, colName2+"9", name+"\n(Pos)")
			colAscii += 2
		}

		f.SetCellValue(sheetName, "X8", "Total Wadah")
		f.MergeCell(sheetName, "X8", "Y8")
		f.SetCellValue(sheetName, "X9", "Jml")
		f.SetCellValue(sheetName, "Y9", "Pos")

		f.SetCellValue(sheetName, "Z8", "Status Jentik")
		f.MergeCell(sheetName, "Z8", "Z9")

		f.SetCellStyle(sheetName, "A8", "Z9", headerStyle)

		// --- ISI TABEL (DATA DETAIL PER RUMAH) ---
		startRow := 10
		totalRumahDiperiksa := len(groupedReports[key])
		totalRumahPositif := 0
		totalKontainerSeluruh := 0
		totalKontainerPositif := 0

		for i, r := range groupedReports[key] {
			row := startRow + i
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), r.FamilyHeadName) // Menampilkan Nama KK
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), r.RT)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), r.Latitude)  // Menampilkan Koordinat
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), r.Longitude) // Menampilkan Koordinat

			if r.LarvaeStatus == 1 {
				totalRumahPositif++
				f.SetCellValue(sheetName, fmt.Sprintf("Z%d", row), "Positif Jentik")
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("Z%d", row), "Aman / Negatif")
			}

			// Mapping wadah yang dilaporkan di rumah ini
			cMap := make(map[string]domain.ContainerDetail)
			for _, cd := range r.ContainerDetails {
				if cd.ContainerType != nil {
					cMap[cd.ContainerType.Name] = cd
				}
			}

			rowContainerTotal := 0
			rowContainerPos := 0
			colDataAscii := 70 // Mulai dari kolom 'F'

			// Sebar jumlah wadah ke kolom masing-masing
			for _, cName := range containers {
				cd, exists := cMap[cName]
				colName1 := string(rune(colDataAscii))
				colName2 := string(rune(colDataAscii + 1))

				if exists {
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName1, row), cd.InspectedCount)
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName2, row), cd.PositiveCount)
					rowContainerTotal += cd.InspectedCount
					rowContainerPos += cd.PositiveCount
				} else {
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName1, row), 0)
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName2, row), 0)
				}
				colDataAscii += 2
			}

			// Subtotal Kontainer Rumah
			f.SetCellValue(sheetName, fmt.Sprintf("X%d", row), rowContainerTotal)
			f.SetCellValue(sheetName, fmt.Sprintf("Y%d", row), rowContainerPos)

			totalKontainerSeluruh += rowContainerTotal
			totalKontainerPositif += rowContainerPos
		}

		// --- FOOTER (REKAP KESELURUHAN RW TSB) ---
		footerRow := startRow + len(groupedReports[key]) + 2
		abj, ci := 0.0, 0.0
		if totalRumahDiperiksa > 0 {
			abj = (float64(totalRumahDiperiksa-totalRumahPositif) / float64(totalRumahDiperiksa)) * 100
		}
		if totalKontainerSeluruh > 0 {
			ci = (float64(totalKontainerPositif) / float64(totalKontainerSeluruh)) * 100
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("REKAPITULASI %s - RW %s", key.Village, key.RW))
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

	// Hapus sheet bawaan (Sheet1) yang kosong
	f.DeleteSheet(defaultSheet)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
