package services

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type villageService struct {
	repo ports.VillageRepository
}

func NewVillageService(repo ports.VillageRepository) ports.VillageService {
	return &villageService{
		repo: repo,
	}
}

func (s *villageService) GetAllVillages(ctx context.Context) ([]domain.Village, error) {
	return s.repo.FindAll(ctx)
}

func (s *villageService) GetVillageByID(ctx context.Context, id string) (*domain.Village, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *villageService) GetPaginated(ctx context.Context, page, limit int) ([]domain.Village, int64, error) {
	// Proteksi nilai negatif atau 0
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	// Batas wajar maksimal data desa per request
	if limit > 100 {
		limit = 100
	}

	return s.repo.GetPaginated(ctx, page, limit)
}
