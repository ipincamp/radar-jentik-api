package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/xuri/excelize/v2"
)

type mapCacheEntry struct {
	data      []domain.InspectionReport
	expiresAt time.Time
}

type inspectionReportService struct {
	repo        ports.InspectionReportRepository
	villageRepo ports.VillageRepository

	cacheMu  sync.RWMutex
	mapCache map[string]mapCacheEntry
}

func NewInspectionReportService(repo ports.InspectionReportRepository, villageRepo ports.VillageRepository) ports.InspectionReportService {
	return &inspectionReportService{
		repo:        repo,
		villageRepo: villageRepo,
		mapCache:    make(map[string]mapCacheEntry),
	}
}

func (s *inspectionReportService) invalidateMapCache() {
	s.cacheMu.Lock()
	s.mapCache = make(map[string]mapCacheEntry)
	s.cacheMu.Unlock()
}

func (s *inspectionReportService) CreateReport(ctx context.Context, report *domain.InspectionReport) error {
	if len(report.ContainerDetails) == 0 {
		return errors.New("minimal harus melaporkan satu jenis wadah")
	}

	if report.VillageID == "" {
		village, err := s.villageRepo.GetByCoordinate(ctx, report.Latitude, report.Longitude)
		if err != nil {
			return errors.New("lokasi anda berada di luar wilayah, silakan pilih desa secara manual")
		}
		report.VillageID = village.ID
	}

	report.LarvaeStatus = 0
	for _, container := range report.ContainerDetails {
		if container.PositiveCount > 0 {
			report.LarvaeStatus = 1
			break
		}
	}

	if report.InspectedAt.IsZero() {
		report.InspectedAt = time.Now()
	}

	report.ValidationStatus = "pending"

	if err := s.repo.Create(ctx, report); err != nil {
		return err
	}

	s.invalidateMapCache()
	return nil
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
	if err := s.repo.UpdateStatus(ctx, id, status, nil); err != nil {
		return err
	}
	s.invalidateMapCache()
	return nil
}

func (s *inspectionReportService) UpdateStatus(ctx context.Context, reportID string, status string, rejectionReason *string) error {
	var reasonPtr *string

	if status == "reject" {
		if rejectionReason == nil || *rejectionReason == "" {
			return errors.New("alasan penolakan wajib diisi jika menolak laporan")
		}
		reasonPtr = rejectionReason
	} else if status == "accept" {
		reasonPtr = nil
	} else {
		return errors.New("status tidak valid")
	}

	if err := s.repo.UpdateStatus(ctx, reportID, status, reasonPtr); err != nil {
		return err
	}
	s.invalidateMapCache()
	return nil
}

func (s *inspectionReportService) GetMapData(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error) {
	cacheKey := role
	if role == "cadre" {
		cacheKey = "cadre:" + userID
	}

	// 1. Cek cache memori terlebih dahulu
	s.cacheMu.RLock()
	entry, exists := s.mapCache[cacheKey]
	s.cacheMu.RUnlock()

	if exists && time.Now().Before(entry.expiresAt) {
		return entry.data, nil
	}

	// 2. Ambil dari database jika cache kosong/kedaluwarsa
	reports, err := s.repo.GetValidReports(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	// 3. Simpan ke cache selama 2 menit
	s.cacheMu.Lock()
	s.mapCache[cacheKey] = mapCacheEntry{
		data:      reports,
		expiresAt: time.Now().Add(2 * time.Minute),
	}
	s.cacheMu.Unlock()

	return reports, nil
}

func (s *inspectionReportService) ExportToExcel(ctx context.Context, userID, role string) ([]byte, string, error) {
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
		return buf.Bytes(), fmt.Sprintf("RadarJentik_%s_Kosong.xlsx", timeStr), nil
	}

	defaultSheet := f.GetSheetName(f.GetActiveSheetIndex())

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

	for _, villageName := range orderedVillages {
		sheetName := villageName
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		f.NewSheet(sheetName)

		f.SetCellValue(sheetName, "A1", "Kabupaten")
		f.SetCellValue(sheetName, "C1", ": Banyumas")
		f.SetCellValue(sheetName, "A2", "Kecamatan")
		f.SetCellValue(sheetName, "C2", ": Cilongok")
		f.SetCellValue(sheetName, "A3", "Kelurahan/Desa")
		f.SetCellValue(sheetName, "C3", ": "+villageName)
		f.SetCellValue(sheetName, "A4", "Tanggal Unduh")
		f.SetCellValue(sheetName, "C4", ": "+time.Now().Format("02 Jan 2006, 15:04 WIB"))

		f.SetCellValue(sheetName, "A7", "No")
		f.MergeCell(sheetName, "A7", "A8")

		f.SetCellValue(sheetName, "B7", "Nama Kepala Keluarga")
		f.MergeCell(sheetName, "B7", "B8")

		f.SetCellValue(sheetName, "C7", "RW")
		f.MergeCell(sheetName, "C7", "C8")

		f.SetCellValue(sheetName, "D7", "RT")
		f.MergeCell(sheetName, "D7", "D8")

		f.SetCellValue(sheetName, "E7", "Waktu Inspeksi")
		f.MergeCell(sheetName, "E7", "E8")

		f.SetCellValue(sheetName, "F7", "Latitude")
		f.MergeCell(sheetName, "F7", "F8")

		f.SetCellValue(sheetName, "G7", "Longitude")
		f.MergeCell(sheetName, "G7", "G8")

		f.SetCellValue(sheetName, "H7", "Container (Wadah)")
		f.MergeCell(sheetName, "H7", "Y7")

		colIdx := 7
		for _, name := range containers {
			colName1 := excelCols[colIdx]
			colName2 := excelCols[colIdx+1]
			f.SetCellValue(sheetName, colName1+"8", name+"\n(Jml)")
			f.SetCellValue(sheetName, colName2+"8", name+"\n(Pos)")
			colIdx += 2
		}

		f.SetCellValue(sheetName, "Z7", "Total Wadah")
		f.MergeCell(sheetName, "Z7", "AA7")
		f.SetCellValue(sheetName, "Z8", "Jml")
		f.SetCellValue(sheetName, "AA8", "Pos")

		f.SetCellValue(sheetName, "AB7", "Status Jentik")
		f.MergeCell(sheetName, "AB7", "AB8")

		f.SetCellStyle(sheetName, "A7", "AB8", headerStyle)

		startRow := 9
		totalRumahDiperiksa := len(groupedReports[villageName])
		totalRumahPositif := 0
		totalKontainerSeluruh := 0
		totalKontainerPositif := 0

		type ContainerAgg struct {
			Inspected int
			Positive  int
		}

		for i, r := range groupedReports[villageName] {
			row := startRow + i
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), r.FamilyHeadName)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), r.RW)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), r.RT)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), r.CreatedAt.Format("02 Jan 2006, 15:04 WIB"))
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), r.Latitude)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), r.Longitude)

			if r.LarvaeStatus == 1 {
				totalRumahPositif++
				f.SetCellValue(sheetName, fmt.Sprintf("AB%d", row), "Positif")
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("AB%d", row), "Negatif")
			}

			cMap := make(map[string]ContainerAgg, len(r.ContainerDetails))
			for _, cd := range r.ContainerDetails {
				if cd.ContainerType != nil {
					typeName := cd.ContainerType.Name
					currentAgg := cMap[typeName]
					currentAgg.Inspected += cd.InspectedCount
					currentAgg.Positive += cd.PositiveCount
					cMap[typeName] = currentAgg
				}
			}

			rowContainerTotal := 0
			rowContainerPos := 0
			colDataIdx := 7

			for _, cName := range containers {
				agg, exists := cMap[cName]
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
	villageNameForFile := "Semua_Desa"
	if len(orderedVillages) == 1 {
		villageNameForFile = "Desa_" + strings.ReplaceAll(orderedVillages[0], " ", "_")
	}

	fileName := fmt.Sprintf("RadarJentik_%s_%s.xlsx", timeStr, villageNameForFile)

	return buf.Bytes(), fileName, nil
}

func (s *inspectionReportService) GetPaginatedHistory(ctx context.Context, userID string, page, limit int, search, rt, rw, villageID, date string) ([]domain.InspectionReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetPaginatedHistory(ctx, userID, page, limit, search, rt, rw, villageID, date)
}

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

func (s *inspectionReportService) CreateBulkReport(ctx context.Context, reports []*domain.InspectionReport) error {
	for i, report := range reports {
		if len(report.ContainerDetails) == 0 {
			return fmt.Errorf("laporan ke-%d: minimal harus melaporkan satu jenis wadah", i+1)
		}

		if report.VillageID == "" {
			village, err := s.villageRepo.GetByCoordinate(ctx, report.Latitude, report.Longitude)
			if err != nil {
				return fmt.Errorf("laporan ke-%d: lokasi berada di luar wilayah terdaftar", i+1)
			}
			report.VillageID = village.ID
		}

		report.LarvaeStatus = 0
		for _, container := range report.ContainerDetails {
			if container.PositiveCount > 0 {
				report.LarvaeStatus = 1
				break
			}
		}

		if report.InspectedAt.IsZero() {
			report.InspectedAt = time.Now()
		}

		report.ValidationStatus = "pending"
	}

	if err := s.repo.CreateBulk(ctx, reports); err != nil {
		return err
	}

	s.invalidateMapCache()
	return nil
}

func (s *inspectionReportService) BulkValidateReports(ctx context.Context, reportIDs []string, status string, rejectionReason *string) error {
	if status != "accept" && status != "reject" {
		return errors.New("status validasi tidak dikenali")
	}

	if err := s.repo.BulkValidateReports(ctx, reportIDs, status, rejectionReason); err != nil {
		return err
	}

	s.invalidateMapCache()
	return nil
}

func (s *inspectionReportService) UpdateReport(ctx context.Context, reportID string, report *domain.InspectionReport) error {
	if len(report.ContainerDetails) == 0 {
		return errors.New("minimal harus melaporkan satu jenis wadah")
	}

	if report.VillageID == "" {
		village, err := s.villageRepo.GetByCoordinate(ctx, report.Latitude, report.Longitude)
		if err != nil {
			return errors.New("lokasi anda berada di luar wilayah, silakan pilih desa secara manual")
		}
		report.VillageID = village.ID
	}

	report.LarvaeStatus = 0
	for _, container := range report.ContainerDetails {
		if container.PositiveCount > 0 {
			report.LarvaeStatus = 1
			break
		}
	}

	if report.InspectedAt.IsZero() {
		report.InspectedAt = time.Now()
	}

	if err := s.repo.Update(ctx, reportID, report); err != nil {
		return err
	}

	s.invalidateMapCache()
	return nil
}
