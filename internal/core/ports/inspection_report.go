package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type InspectionReportRepository interface {
	Create(ctx context.Context, report *domain.InspectionReport) error
	GetByUserID(ctx context.Context, userID string) ([]domain.InspectionReport, error)
	GetPending(ctx context.Context) ([]domain.InspectionReport, error)
	UpdateStatus(ctx context.Context, reportID string, status string, rejectionReason *string) error
	GetValidReports(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error)
	GetRecapData(ctx context.Context, userID, role string) ([]domain.ReportRecap, error)
	GetExportData(ctx context.Context, userID, role string) ([]domain.InspectionReport, error)
	GetPaginatedHistory(ctx context.Context, userID string, page, limit int) ([]domain.InspectionReport, int64, error)
	GetPaginatedPending(ctx context.Context, page, limit int) ([]domain.InspectionReport, int64, error)
	CreateBulk(ctx context.Context, reports []*domain.InspectionReport) error
	BulkValidateReports(ctx context.Context, reportIDs []string, status string) error
}

type InspectionReportService interface {
	CreateReport(ctx context.Context, report *domain.InspectionReport) error
	GetCadreHistory(ctx context.Context, userID string) ([]domain.InspectionReport, error)
	GetPendingReports(ctx context.Context) ([]domain.InspectionReport, error)
	ValidateReport(ctx context.Context, id string, status string) error
	GetMapData(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error)
	ExportToExcel(ctx context.Context, userID, role string) ([]byte, string, error)
	UpdateStatus(ctx context.Context, reportID string, status string, rejectionReason *string) error
	GetPaginatedHistory(ctx context.Context, userID string, page, limit int) ([]domain.InspectionReport, int64, error)
	GetPaginatedPending(ctx context.Context, page, limit int) ([]domain.InspectionReport, int64, error)
	CreateBulkReport(ctx context.Context, reports []*domain.InspectionReport) error
	BulkValidateReports(ctx context.Context, reportIDs []string, status string) error
}
