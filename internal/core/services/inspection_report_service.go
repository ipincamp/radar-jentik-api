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
