package services

import (
	"context"
	"math"
	"runtime"
	"sync"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"github.com/ipincamp/radar-jentik-api/pkg/spatial"
)

type idwService struct{}

func NewIDWService() ports.IDWService {
	return &idwService{}
}

func (s *idwService) CalculateIDWGrid(ctx context.Context, req domain.IDWRequest) ([]domain.GridPoint, error) {
	if len(req.Samples) == 0 || req.Resolution <= 0 {
		return []domain.GridPoint{}, nil
	}

	latSteps := int(math.Ceil((req.MaxLat - req.MinLat) / req.Resolution))
	lonSteps := int(math.Ceil((req.MaxLon - req.MinLon) / req.Resolution))

	totalRows := latSteps + 1
	totalCols := lonSteps + 1
	totalPoints := totalRows * totalCols

	// Pre-allocate slice dengan ukuran pasti untuk mencegah realokasi memori berulang
	grid := make([]domain.GridPoint, totalPoints)

	// Worker pool paralel sesuai jumlah core CPU
	numWorkers := runtime.NumCPU()
	if numWorkers > totalRows {
		numWorkers = totalRows
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	var wg sync.WaitGroup
	rowChan := make(chan int, totalRows)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range rowChan {
				// Cek jika HTTP request dibatalkan client
				select {
				case <-ctx.Done():
					return
				default:
				}

				lat := req.MinLat + (float64(i) * req.Resolution)
				rowOffset := i * totalCols

				for j := 0; j <= lonSteps; j++ {
					lon := req.MinLon + (float64(j) * req.Resolution)
					estimatedValue := s.CalculateSinglePoint(lat, lon, req.Samples, req.Power)

					grid[rowOffset+j] = domain.GridPoint{
						Lat:            lat,
						Lon:            lon,
						EstimatedValue: estimatedValue,
					}
				}
			}
		}()
	}

	for i := 0; i <= latSteps; i++ {
		rowChan <- i
	}
	close(rowChan)
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return grid, nil
}

// Fungsi internal untuk menghitung IDW pada 1 titik target (Z_j)
func (s *idwService) CalculateSinglePoint(
	targetLat, targetLon float64,
	samples []domain.SamplePoint,
	power float64,
) float64 {
	var numerator float64
	var denominator float64

	// Batas cutoff derajat (~5.5 km di sekitar ekuator/Cilongok)
	// Sampel di luar radius ini bobotnya mendekati 0 sehingga dilewati tanpa Haversine
	const maxDegreeCutoff = 0.05
	isPowerTwo := power == 2.0

	for i := range samples {
		sample := &samples[i]

		// 1. Fast Bounding Box Filter (hanya operasi pengurangan float biasa)
		dLat := targetLat - sample.Lat
		if dLat > maxDegreeCutoff || dLat < -maxDegreeCutoff {
			continue
		}
		dLon := targetLon - sample.Lon
		if dLon > maxDegreeCutoff || dLon < -maxDegreeCutoff {
			continue
		}

		// 2. Hitung jarak Haversine hanya untuk sampel yang dekat
		dist := spatial.HaversineDistance(targetLat, targetLon, sample.Lat, sample.Lon)
		if dist < 1e-9 {
			return sample.Value
		}

		// 3. Hindari math.Pow yang mahal jika power == 2 (standar IDW)
		var weight float64
		if isPowerTwo {
			weight = 1.0 / (dist * dist)
		} else {
			weight = 1.0 / math.Pow(dist, power)
		}

		numerator += weight * sample.Value
		denominator += weight
	}

	// Fallback jika tidak ada sampel dalam radius cutoff (gunakan seluruh sampel)
	if denominator == 0 {
		for i := range samples {
			sample := &samples[i]
			dist := spatial.HaversineDistance(targetLat, targetLon, sample.Lat, sample.Lon)
			if dist < 1e-9 {
				return sample.Value
			}
			var weight float64
			if isPowerTwo {
				weight = 1.0 / (dist * dist)
			} else {
				weight = 1.0 / math.Pow(dist, power)
			}
			numerator += weight * sample.Value
			denominator += weight
		}
	}

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}
