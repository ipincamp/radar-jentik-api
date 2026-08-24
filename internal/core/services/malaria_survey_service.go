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

type malariaSurveyService struct {
	repo        ports.MalariaSurveyRepository
	villageRepo ports.VillageRepository
}

func NewMalariaSurveyService(repo ports.MalariaSurveyRepository, villageRepo ports.VillageRepository) ports.MalariaSurveyService {
	return &malariaSurveyService{
		repo:        repo,
		villageRepo: villageRepo,
	}
}

func (s *malariaSurveyService) CreateSurvey(ctx context.Context, survey *domain.MalariaSurvey) error {
	// Auto-deteksi Desa berdasarkan Geografis/Koordinat (Jika VillageID kosong)
	if survey.VillageID == "" {
		village, err := s.villageRepo.GetByCoordinate(ctx, survey.Latitude, survey.Longitude)
		if err != nil {
			return errors.New("lokasi berada di luar area operasional desa, harap pilih desa secara manual")
		}
		survey.VillageID = village.ID
	}

	// Hitung Kepadatan Larva jika ditemukan jentik (Kepadatan = Jumlah Larva / Jumlah Cidukan)
	if survey.IsLarvaeFound {
		if survey.ScoopCount <= 0 {
			return errors.New("jika ditemukan larva, jumlah cidukan tidak boleh 0")
		}
		// Konversi ke float64 untuk pembagian presisi desimal
		survey.LarvaeDensity = float64(survey.LarvaeCount) / float64(survey.ScoopCount)
	} else {
		// Pastikan data ini direset jika kader menekan "Tidak"
		survey.LarvaeSpecies = ""
		survey.ScoopCount = 0
		survey.LarvaeCount = 0
		survey.LarvaeDensity = 0
	}

	if survey.InspectedAt.IsZero() {
		survey.InspectedAt = time.Now()
	}

	return s.repo.Create(ctx, survey)
}

func (s *malariaSurveyService) GetPaginatedHistory(ctx context.Context, userID string, page, limit int) ([]domain.MalariaSurvey, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return s.repo.GetPaginatedHistory(ctx, userID, page, limit)
}

func (s *malariaSurveyService) UpdateSurvey(ctx context.Context, id string, survey *domain.MalariaSurvey) error {
	// Hitung ulang kepadatan larva
	if survey.IsLarvaeFound {
		if survey.ScoopCount > 0 {
			survey.LarvaeDensity = float64(survey.LarvaeCount) / float64(survey.ScoopCount)
		}
	} else {
		survey.LarvaeSpecies = ""
		survey.ScoopCount = 0
		survey.LarvaeCount = 0
		survey.LarvaeDensity = 0
	}

	if survey.InspectedAt.IsZero() {
		survey.InspectedAt = time.Now()
	}

	return s.repo.Update(ctx, id, survey)
}

func (s *malariaSurveyService) ExportToExcel(ctx context.Context, userID, role string) ([]byte, string, error) {
	surveys, err := s.repo.GetExportData(ctx, userID, role)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer f.Close()
	timeStr := time.Now().Format("02Jan2006_1504")

	if len(surveys) == 0 {
		var buf bytes.Buffer
		f.Write(&buf)
		return buf.Bytes(), fmt.Sprintf("SurveiMalaria_%s_Kosong.xlsx", timeStr), nil
	}

	defaultSheet := f.GetSheetName(f.GetActiveSheetIndex())
	groupedSurveys := make(map[string][]domain.MalariaSurvey)
	var orderedVillages []string

	// Kelompokkan data berdasarkan Desa
	for _, srv := range surveys {
		vName := srv.Village.Name
		if _, exists := groupedSurveys[vName]; !exists {
			orderedVillages = append(orderedVillages, vName)
		}
		groupedSurveys[vName] = append(groupedSurveys[vName], srv)
	}

	// Style untuk Header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Kamus terjemahan hari ke Bahasa Indonesia
	hariIndo := map[string]string{
		"Sunday":    "Minggu",
		"Monday":    "Senin",
		"Tuesday":   "Selasa",
		"Wednesday": "Rabu",
		"Thursday":  "Kamis",
		"Friday":    "Jumat",
		"Saturday":  "Sabtu",
	}

	// Buat Sheet untuk setiap desa
	for _, vName := range orderedVillages {
		sheetName := vName
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		f.NewSheet(sheetName)

		// Mengambil tanggal inspeksi dari data pertama di desa ini
		firstSurveyDate := groupedSurveys[vName][0].InspectedAt
		namaHari := hariIndo[firstSurveyDate.Format("Monday")]
		tanggalFormat := firstSurveyDate.Format("02 Jan 2006")
		hariTanggalStr := fmt.Sprintf("Hari/Tanggal : %s, %s", namaHari, tanggalFormat)

		// --- KOP SURVEI ---
		f.SetCellValue(sheetName, "A1", "FORMULIR SURVEI TEMPAT PERKEMBANGBIAKAN Anopheles spp.")
		f.SetCellValue(sheetName, "A2", "DI WILAYAH KERJA PUSKESMAS CILONGOK II")
		f.SetCellValue(sheetName, "A3", "Desa : "+vName)
		f.SetCellValue(sheetName, "A4", hariTanggalStr) // -> Diubah menggunakan tanggal survei

		// --- HEADER TABEL BARIS 1 & 2 (MERGE CELLS) ---

		// Kolom A-F: Merge baris 6 dan 7 vertikal
		headersAF := []string{"No", "Jam", "Dusun/Grumbul", "RT", "RW", "Tipe Tempat Perindukan\n(sawah/ mata air/ sungai /parit/ lagoon dll)"}
		colsAF := []string{"A", "B", "C", "D", "E", "F"}
		for i, col := range colsAF {
			f.SetCellValue(sheetName, col+"6", headersAF[i])
			f.MergeCell(sheetName, col+"6", col+"7")
		}

		// Titik Koordinat (Merge G6:H6)
		f.SetCellValue(sheetName, "G6", "Titik Koordinat")
		f.MergeCell(sheetName, "G6", "H6")
		f.SetCellValue(sheetName, "G7", "longitude")
		f.SetCellValue(sheetName, "H7", "latitude")

		// Karakteristik Tempat Perindukan (Merge I6:M6)
		f.SetCellValue(sheetName, "I6", "Karakteristik Tempat Perindukan")
		f.MergeCell(sheetName, "I6", "M6")
		f.SetCellValue(sheetName, "I7", "Fisik\n(Pencahayaan (Naungan pohon/ teduh/ terkena sinar matahari), aliran air)")
		f.SetCellValue(sheetName, "J7", "Biologi\n( Adanya tanaman (tanaman air/ alga/ lumut) atau hewan)")
		f.SetCellValue(sheetName, "K7", "Kimia\n(Kadar garam (salinitas), PH, suhu air)")
		f.SetCellValue(sheetName, "L7", "Luas\nTempat\nPerindukan")
		f.SetCellValue(sheetName, "M7", "Kedalaman\nTempat\nPerindukan")

		// Ditemukan Jentik (Merge N6:P6)
		f.SetCellValue(sheetName, "N6", "Ditemukan Jentik")
		f.MergeCell(sheetName, "N6", "P6")
		f.SetCellValue(sheetName, "N7", "Tidak")
		f.SetCellValue(sheetName, "O7", "Ya")
		f.SetCellValue(sheetName, "P7", "Spesies larva\n(Anopheles/ aedes/ culex/ dll)") // Jika ditemukan jentik, maka wajib diisi spesies larva

		// Kepadatan (Merge Q6:S6)
		f.SetCellValue(sheetName, "Q6", "Kepadatan")
		f.MergeCell(sheetName, "Q6", "S6")
		f.SetCellValue(sheetName, "Q7", "Jumlah\nCidukan")
		f.SetCellValue(sheetName, "R7", "Jumlah Larva\n1x Cidukan")
		f.SetCellValue(sheetName, "S7", "Kepadatan\n(Jumlah Larva : Jumlah Cidukan)")

		// Terapkan Styling Kotak/Garis Header untuk area A6 sampai S7
		f.SetCellStyle(sheetName, "A6", "S7", headerStyle)
		f.SetColWidth(sheetName, "P", "S", 18)
		f.SetColWidth(sheetName, "I", "M", 15)

		// --- ISI DATA TABEL ---
		rowIdx := 8 // Data dimulai dari baris ke-8 karena baris 6 & 7 dipakai header
		for i, srv := range groupedSurveys[vName] {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), i+1)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), srv.InspectedAt.Format("15:04")) // Format Jam
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), srv.Dusun)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIdx), srv.RT)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIdx), srv.RW)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIdx), srv.BreedingPlaceType)

			// Koordinat
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIdx), srv.Longitude) // longitude
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIdx), srv.Latitude)  // latitude

			// Format rincian karakteristik habitat
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIdx), fmt.Sprintf("%s, %s", srv.PhysicalLighting, srv.PhysicalWaterFlow))
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowIdx), fmt.Sprintf("%s, %s", srv.BioPlants, srv.BioAnimals))
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIdx), fmt.Sprintf("Sal:%.2f, pH:%.2f, Suhu:%.2f", srv.ChemSalinity, srv.ChemPH, srv.ChemWaterTemp))
			f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowIdx), srv.Area)
			f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowIdx), srv.Depth)

			// Pengecekan status jentik: Kolom N = Tidak, Kolom O = Ya
			if srv.IsLarvaeFound {
				f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowIdx), "")  // Tidak
				f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowIdx), "V") // Ya
				f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowIdx), srv.LarvaeSpecies)
				f.SetCellValue(sheetName, fmt.Sprintf("Q%d", rowIdx), srv.ScoopCount)
				f.SetCellValue(sheetName, fmt.Sprintf("R%d", rowIdx), srv.LarvaeCount)
				f.SetCellValue(sheetName, fmt.Sprintf("S%d", rowIdx), srv.LarvaeDensity)
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowIdx), "V") // Tidak
				f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowIdx), "")  // Ya
				f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowIdx), "-")
				f.SetCellValue(sheetName, fmt.Sprintf("Q%d", rowIdx), "-")
				f.SetCellValue(sheetName, fmt.Sprintf("R%d", rowIdx), "-")
				f.SetCellValue(sheetName, fmt.Sprintf("S%d", rowIdx), "-")
			}

			// Tambahkan border style pada baris data
			borderStyle, _ := f.NewStyle(&excelize.Style{
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
				},
				Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
			})
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("S%d", rowIdx), borderStyle)

			rowIdx++
		}
	}

	f.DeleteSheet(defaultSheet)
	var buf bytes.Buffer
	f.Write(&buf)

	fileName := fmt.Sprintf("SurveiMalaria_%s.xlsx", timeStr)
	return buf.Bytes(), fileName, nil
}
