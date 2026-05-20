package services

import (
	"context"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type inspectionReportService struct {
	repo ports.InspectionReportRepository
}

func NewInspectionReportService(repo ports.InspectionReportRepository) ports.InspectionReportService {
	return &inspectionReportService{
		repo: repo,
	}
}

// CreateReport bertugas memproses payload dan mengirimkannya ke repository
func (s *inspectionReportService) CreateReport(ctx context.Context, report *domain.InspectionReport) error {
	// Anda bisa menambahkan logika bisnis di sini (misal: validasi tambahan, perhitungan, dll)

	// Set default value sebelum masuk ke database
	report.ValidationStatus = "pending"
	report.InspectedAt = time.Now()

	// Memanggil repository yang sudah membungkus proses insert ke dalam Transaksi SQL
	err := s.repo.Create(ctx, report)
	if err != nil {
		return err
	}

	return nil
}

func (s *inspectionReportService) GetCadreHistory(ctx context.Context, userID string) ([]domain.InspectionReport, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *inspectionReportService) GetPendingReports(ctx context.Context) ([]domain.InspectionReport, error) {
	return s.repo.GetPending(ctx)
}

func (s *inspectionReportService) ValidateReport(ctx context.Context, id string, status string) error {
	// Memastikan status yang masuk hanya 'accept' atau 'reject'
	if status != "accept" && status != "reject" {
		return context.DeadlineExceeded // atau buat custom error
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *inspectionReportService) GetMapData(ctx context.Context) ([]domain.InspectionReport, error) {
	return s.repo.GetValidReports(ctx)
}
