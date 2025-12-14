package services

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type ReportService struct {
	repo ports.ReportRepository
}

func NewReportService(repo ports.ReportRepository) ports.ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) Create(ctx context.Context, reporterID string, req ports.CreateReportRequest) error {
	newReport := &domain.Report{
		ReporterID:         reporterID,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		LarvaeDensityIndex: req.LarvaeDensityIndex,
		PhotoURL:           req.PhotoURL,
		Notes:              req.Notes,
		Status:             "pending",
	}

	return s.repo.Save(ctx, newReport)
}
