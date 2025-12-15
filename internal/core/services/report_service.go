package services

import (
	"context"
	"math"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type ReportService struct {
	repo ports.ReportRepository
}

// calculateIDWValue menghitung nilai Z' pada satu titik grid (S0)
func calculateIDWValue(gridLat, gridLon float64, samples []*domain.Report, p float64) float64 {
	var numerator float64 = 0   // Pembilang (Sigma lambda * Z)
	var denominator float64 = 0 // Penyebut (Sigma lambda)

	for _, sample := range samples {
		// Hitung Jarak Euclidean (Persamaan 3)
		// Catatan: Untuk presisi tinggi geospasial bisa pakai Haversine,
		// tapi Euclidean cukup untuk skala lokal kecil sesuai proposal.
		dist := math.Sqrt(math.Pow(gridLat-sample.Latitude, 2) + math.Pow(gridLon-sample.Longitude, 2))

		// Handle Singularity: Jika jarak sangat dekat (titik grid tepat di atas sampel)
		if dist < 0.000001 {
			return float64(sample.LarvaeDensityIndex) // Atau gunakan status 1/0
		}

		// Hitung Bobot (Weight) = 1 / d^p
		weight := 1.0 / math.Pow(dist, p)

		// Nilai Z (Nilai Sampel)
		// Menggunakan LarvaeDensityIndex. Jika ingin biner (0/1), ubah logika ini.
		zValue := float64(sample.LarvaeDensityIndex)

		numerator += weight * zValue
		denominator += weight
	}

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
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

func (s *ReportService) Validate(ctx context.Context, reportID, verifierID string, req ports.ValidateReportRequest) error {
	// 1. Cari laporan berdasarkan ID
	report, err := s.repo.FindByID(ctx, reportID)
	if err != nil {
		return err // Bisa dikembangkan dengan custom error "Report Not Found"
	}

	// 2. Update status dan data verifikator
	now := time.Now()
	report.Status = req.Status
	report.Notes = req.Notes
	report.VerifierID = &verifierID
	report.VerifiedAt = &now

	// 3. Simpan perubahan
	return s.repo.Update(ctx, report)
}

func (s *ReportService) GetHeatmapData(ctx context.Context, req ports.GetHeatmapRequest) ([]ports.HeatmapPoint, error) {
	// 1. Set Default Parameter
	if req.PowerParameter == 0 {
		req.PowerParameter = 2.0 // Default p = 2
	}
	if req.GridResolution == 0 {
		req.GridResolution = 0.005 // Default resolusi grid (sekitar ~500m)
	}

	// 2. Ambil SEMUA data laporan dengan status 'verified'
	// Catatan: Untuk MVP masih ambil sampel data cukup banyak, misal limit 1000
	reports, _, err := s.repo.FindAll(ctx, 1, 1000)
	if err != nil {
		return nil, err
	}

	if len(reports) == 0 {
		return []ports.HeatmapPoint{}, nil
	}

	// 3. Tentukan Bounding Box (Batas Wilayah)
	minLat, maxLat := reports[0].Latitude, reports[0].Latitude
	minLon, maxLon := reports[0].Longitude, reports[0].Longitude

	for _, r := range reports {
		if r.Latitude < minLat {
			minLat = r.Latitude
		}
		if r.Latitude > maxLat {
			maxLat = r.Latitude
		}
		if r.Longitude < minLon {
			minLon = r.Longitude
		}
		if r.Longitude > maxLon {
			maxLon = r.Longitude
		}
	}

	// Tambahkan sedikit buffer di tepi peta
	buffer := 0.01
	minLat -= buffer
	maxLat += buffer
	minLon -= buffer
	maxLon += buffer

	// 4. Generate Grid & Kalkulasi IDW
	var heatmap []ports.HeatmapPoint

	// Loop Latitude (Y)
	for lat := minLat; lat <= maxLat; lat += req.GridResolution {
		// Loop Longitude (X)
		for lon := minLon; lon <= maxLon; lon += req.GridResolution {
			estimatedValue := calculateIDWValue(lat, lon, reports, req.PowerParameter)

			// Hanya masukkan ke hasil jika nilainya signifikan (opsional)
			if estimatedValue > 0 {
				heatmap = append(heatmap, ports.HeatmapPoint{
					Latitude:  lat,
					Longitude: lon,
					Value:     estimatedValue,
				})
			}
		}
	}

	return heatmap, nil
}
