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

type ValidateReportRequest struct {
	Status string `json:"status" validate:"required,oneof=verified rejected"`
	Notes  string `json:"notes"`
}

// Struct untuk parameter Pagination
type FindReportsRequest struct {
	Page          int    `query:"page"`
	Limit         int    `query:"limit"`
	RequestorID   string `json:"-"` // Diisi oleh Handler (Internal)
	RequestorRole string `json:"-"` // Diisi oleh Handler (Internal)
}

// Request untuk generate heatmap
type GetHeatmapRequest struct {
	PowerParameter float64 `query:"p" json:"p"`     // Default 2.0
	GridResolution float64 `query:"res" json:"res"` // Jarak antar titik grid (derajat), misal 0.001
}

// Struct untuk response yang menyertakan metadata pagination
type PaginatedResponse struct {
	Data     []*domain.Report `json:"data"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	LastPage int              `json:"last_page"`
}

// Representasi satu titik hasil interpolasi (x, y, z')
type HeatmapPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Value     float64 `json:"value"` // Nilai estimasi risiko (0.0 - 1.0 atau lebih)
}

type ReportRepository interface {
	Save(ctx context.Context, report *domain.Report) error
	FindAll(ctx context.Context, page, limit int, filterUserID *string) ([]*domain.Report, int64, error)
	FindByID(ctx context.Context, id string) (*domain.Report, error)
	Update(ctx context.Context, report *domain.Report) error
}

type ReportService interface {
	Create(ctx context.Context, reporterID string, req CreateReportRequest) error
	GetAll(ctx context.Context, req FindReportsRequest) (*PaginatedResponse, error)
	Validate(ctx context.Context, reportID, verifierID string, req ValidateReportRequest) error
	GetHeatmapData(ctx context.Context, req GetHeatmapRequest) ([]HeatmapPoint, error)
}
