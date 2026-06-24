package services

import (
	"context"
	"math"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/spatial"
)

type idwService struct{}

func NewIDWService() ports.IDWService {
	return &idwService{}
}

func (s *idwService) CalculateIDWGrid(ctx context.Context, req domain.IDWRequest) ([]domain.GridPoint, error) {
	var grid []domain.GridPoint

	// 1. Hitung jumlah langkah (steps) berbentuk integer
	latSteps := int(math.Ceil((req.MaxLat - req.MinLat) / req.Resolution))
	lonSteps := int(math.Ceil((req.MaxLon - req.MinLon) / req.Resolution))

	// 2. Looping menggunakan index (integer) lalu dikalikan dengan resolusi
	// Ini menjamin jarak (lat, lon) sama persis dan tidak bergeser nilainya
	for i := 0; i <= latSteps; i++ {
		lat := req.MinLat + (float64(i) * req.Resolution)

		for j := 0; j <= lonSteps; j++ {
			lon := req.MinLon + (float64(j) * req.Resolution)

			// 3. Hitung estimasi IDW untuk titik (lat, lon) ini
			estimatedValue := s.CalculateSinglePoint(lat, lon, req.Samples, req.Power)

			grid = append(grid, domain.GridPoint{
				Lat:            lat,
				Lon:            lon,
				EstimatedValue: estimatedValue,
			})
		}
	}

	return grid, nil
}

// Fungsi internal untuk menghitung IDW pada 1 titik target (Z_j)
func (s *idwService) CalculateSinglePoint(
	targetLat, targetLon float64,
	samples []domain.SamplePoint,
	power float64,
) float64 {
	var numerator float64 = 0   // Pembilang: Sum(Zi / d^p)
	var denominator float64 = 0 // Penyebut: Sum(1 / d^p)

	for _, sample := range samples {
		// Hitung jarak (d) menggunakan Haversine
		dist := spatial.HaversineDistance(targetLat, targetLon, sample.Lat, sample.Lon)

		// Jika titik target tepat berada di atas titik sampel, nilainya sama persis
		if dist == 0 {
			return sample.Value
		}

		// Hitung bobot (1 / d^p)
		weight := 1.0 / math.Pow(dist, power)

		numerator += weight * sample.Value
		denominator += weight
	}

	if denominator == 0 {
		return 0
	}

	// Hasil Z_j
	return numerator / denominator
}
