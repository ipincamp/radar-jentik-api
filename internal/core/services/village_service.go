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
