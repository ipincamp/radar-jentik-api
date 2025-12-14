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

// Struct untuk parameter Pagination
type FindReportsRequest struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

// Struct untuk response yang menyertakan metadata pagination
type PaginatedResponse struct {
	Data     []*domain.Report `json:"data"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	LastPage int              `json:"last_page"`
}

type ReportRepository interface {
	Save(ctx context.Context, report *domain.Report) error
	FindAll(ctx context.Context, page, limit int) ([]*domain.Report, int64, error)
}

type ReportService interface {
	Create(ctx context.Context, reporterID string, req CreateReportRequest) error
	GetAll(ctx context.Context, req FindReportsRequest) (*PaginatedResponse, error)
}
