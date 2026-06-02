package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type IDWService interface {
	CalculateIDWGrid(ctx context.Context, req domain.IDWRequest) ([]domain.GridPoint, error)
	CalculateSinglePoint(targetLat, targetLon float64, samples []domain.SamplePoint, power float64) float64
}
