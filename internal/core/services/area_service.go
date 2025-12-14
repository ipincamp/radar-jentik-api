package services

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type AreaService struct {
	repo ports.AreaRepository
}

func NewAreaService(repo ports.AreaRepository) ports.AreaService {
	return &AreaService{repo: repo}
}

func (s *AreaService) GetAllAreas(ctx context.Context) ([]*domain.Area, error) {
	return s.repo.FindAll(ctx)
}

// ImportGeoJSON untuk kebutuhan upload via API di masa depan (opsional)
func (s *AreaService) ImportGeoJSON(ctx context.Context, rawData []byte) error {
	// Implementasi logika parsing jika fitur upload file via API dibutuhkan nanti
	return nil
}
