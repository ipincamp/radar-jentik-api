package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type InspectionReportRepository interface {
	Create(ctx context.Context, report *domain.InspectionReport) error
	GetByUserID(ctx context.Context, userID string) ([]domain.InspectionReport, error)
	GetPending(ctx context.Context) ([]domain.InspectionReport, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	GetValidReports(ctx context.Context) ([]domain.InspectionReport, error) // Untuk Peta IDW
}

type InspectionReportService interface {
	CreateReport(ctx context.Context, report *domain.InspectionReport) error
	GetCadreHistory(ctx context.Context, userID string) ([]domain.InspectionReport, error)
	GetPendingReports(ctx context.Context) ([]domain.InspectionReport, error)
	ValidateReport(ctx context.Context, id string, status string) error
	GetMapData(ctx context.Context) ([]domain.InspectionReport, error)
}
