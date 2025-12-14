package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type AreaRepository interface {
	Save(ctx context.Context, area *domain.Area, geoJSONGeometry string) error
	FindAll(ctx context.Context) ([]*domain.Area, error)
}

type AreaService interface {
	ImportGeoJSON(ctx context.Context, rawData []byte) error
	GetAllAreas(ctx context.Context) ([]*domain.Area, error)
}
