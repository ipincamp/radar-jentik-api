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

func (s *ReportService) GetAll(ctx context.Context, req ports.FindReportsRequest) (*ports.PaginatedResponse, error) {
	// Default Value jika tidak diisi
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10 // Default 10 item per halaman
	}
	// Limit maksimal untuk mencegah load berlebih
	if req.Limit > 100 {
		req.Limit = 100
	}

	reports, total, err := s.repo.FindAll(ctx, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	// Hitung Last Page
	lastPage := int(total) / req.Limit
	if int(total)%req.Limit != 0 {
		lastPage++
	}

	return &ports.PaginatedResponse{
		Data:     reports,
		Total:    total,
		Page:     req.Page,
		LastPage: lastPage,
	}, nil
}
