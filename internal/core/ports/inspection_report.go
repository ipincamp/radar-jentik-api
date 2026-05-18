package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type InspectionReportRepository interface {
	Create(ctx context.Context, report *domain.InspectionReport) error
}

type InspectionReportService interface {
	CreateReport(ctx context.Context, report *domain.InspectionReport) error
}
