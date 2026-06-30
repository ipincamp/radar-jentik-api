package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type VillageRepository interface {
	FindAll(ctx context.Context) ([]domain.Village, error)
	FindByID(ctx context.Context, id string) (*domain.Village, error)
	Create(ctx context.Context, village *domain.Village) error
	GetPaginated(ctx context.Context, page, limit int) ([]domain.Village, int64, error)
}

type VillageService interface {
	GetAllVillages(ctx context.Context) ([]domain.Village, error)
	GetVillageByID(ctx context.Context, id string) (*domain.Village, error)
	GetPaginated(ctx context.Context, page, limit int) ([]domain.Village, int64, error)
}
