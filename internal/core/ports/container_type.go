package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type ContainerTypeRepository interface {
	FindAllActive(ctx context.Context) ([]domain.ContainerType, error)
}

type ContainerTypeService interface {
	GetActiveTypes(ctx context.Context) ([]domain.ContainerType, error)
}
