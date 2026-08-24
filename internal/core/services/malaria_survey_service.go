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

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Buat Sheet untuk setiap desa
	for _, vName := range orderedVillages {
		sheetName := vName
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		f.NewSheet(sheetName)

		// --- KOP SURVEI ---
		f.SetCellValue(sheetName, "A1", "FORMULIR SURVEI TEMPAT PERKEMBANGBIAKAN Anopheles spp.")
		f.SetCellValue(sheetName, "A2", "DI WILAYAH KERJA PUSKESMAS CILONGOK II")
		f.SetCellValue(sheetName, "A3", "Desa : "+vName)
		f.SetCellValue(sheetName, "A4", "Tanggal Unduh : "+time.Now().Format("02 Jan 2006"))

		// --- HEADER TABEL SESUAI FORM PDF ---
		f.SetCellValue(sheetName, "A6", "No.")
		f.SetCellValue(sheetName, "B6", "Jam")
		f.SetCellValue(sheetName, "C6", "Dusun/Grumbul")
		f.SetCellValue(sheetName, "D6", "RT/RW")
		f.SetCellValue(sheetName, "E6", "Tipe Tempat Perindukan")
		f.SetCellValue(sheetName, "F6", "Titik koordinat")
		f.SetCellValue(sheetName, "G6", "Fisik (Pencahayaan, Aliran Air)")
		f.SetCellValue(sheetName, "H6", "Biologi (Tanaman, Hewan)")
		f.SetCellValue(sheetName, "I6", "Kimia (Salinitas, pH, Suhu)")
		f.SetCellValue(sheetName, "J6", "Luas Tempat Perindukan (m2)")
		f.SetCellValue(sheetName, "K6", "Kedalaman Tempat Perindukan (m)")
		f.SetCellValue(sheetName, "L6", "Ditemukan Jentik Tidak (V)")
		f.SetCellValue(sheetName, "M6", "Ditemukan Jentik Ya (V)")
		f.SetCellValue(sheetName, "N6", "Spesies larva")
		f.SetCellValue(sheetName, "O6", "Jumlah Cidukan")
		f.SetCellValue(sheetName, "P6", "Jumlah 1x cidukan")
		f.SetCellValue(sheetName, "Q6", "Kepadatan")

		f.SetCellStyle(sheetName, "A6", "Q6", headerStyle)

		// --- ISI DATA TABEL ---
		rowIdx := 7
		for i, srv := range groupedSurveys[vName] {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIdx), i+1)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIdx), srv.InspectedAt.Format("15:04")) // Format Jam
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx), srv.Dusun)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIdx), fmt.Sprintf("RT %s/RW %s", srv.RT, srv.RW))
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIdx), srv.BreedingPlaceType)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIdx), fmt.Sprintf("%f, %f", srv.Latitude, srv.Longitude))
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIdx), fmt.Sprintf("%s, %s", srv.PhysicalLighting, srv.PhysicalWaterFlow))
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIdx), fmt.Sprintf("%s, %s", srv.BioPlants, srv.BioAnimals))
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIdx), fmt.Sprintf("Sal:%f, pH:%f, Suhu:%f", srv.ChemSalinity, srv.ChemPH, srv.ChemWaterTemp))
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowIdx), srv.Area)
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIdx), srv.Depth)

			if srv.IsLarvaeFound {
				f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowIdx), "")  // Tidak
				f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowIdx), "V") // Ya
				f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowIdx), srv.LarvaeSpecies)
				f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowIdx), srv.ScoopCount)
				f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowIdx), srv.LarvaeCount)
				f.SetCellValue(sheetName, fmt.Sprintf("Q%d", rowIdx), srv.LarvaeDensity)
			} else {
				f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowIdx), "V") // Tidak
				f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowIdx), "")  // Ya
				f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowIdx), "-")
				f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowIdx), "-")
				f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowIdx), "-")
				f.SetCellValue(sheetName, fmt.Sprintf("Q%d", rowIdx), "-")
			}
			rowIdx++
		}
	}

	f.DeleteSheet(defaultSheet)
	var buf bytes.Buffer
	f.Write(&buf)

	fileName := fmt.Sprintf("SurveiMalaria_%s.xlsx", timeStr)
	return buf.Bytes(), fileName, nil
}
