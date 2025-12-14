package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type CreateReportRequest struct {
	Latitude           float64 `json:"latitude" validate:"required,latitude"`
	Longitude          float64 `json:"longitude" validate:"required,longitude"`
	LarvaeDensityIndex int     `json:"larvae_density_index" validate:"min=0"`
	PhotoURL           string  `json:"photo_url"`
	Notes              string  `json:"notes"`
}

type ReportRepository interface {
	Save(ctx context.Context, report *domain.Report) error
}

type ReportService interface {
	Create(ctx context.Context, reporterID string, req CreateReportRequest) error
}
