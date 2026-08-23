package services

import (
	"context"
	"errors"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
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
