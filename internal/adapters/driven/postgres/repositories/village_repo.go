package repositories

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"

	"gorm.io/gorm"
)

type villageRepository struct {
	db *gorm.DB
}

func NewVillageRepository(db *gorm.DB) ports.VillageRepository {
	return &villageRepository{
		db: db,
	}
}

func (r *villageRepository) FindAll(ctx context.Context) ([]domain.Village, error) {
	var villages []domain.Village
	// Mengambil semua desa yang tidak dihapus (soft delete), urut berdasarkan nama
	if err := r.db.WithContext(ctx).Order("name asc").Find(&villages).Error; err != nil {
		return nil, err
	}
	return villages, nil
}

func (r *villageRepository) FindByID(ctx context.Context, id string) (*domain.Village, error) {
	var village domain.Village
	if err := r.db.WithContext(ctx).First(&village, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &village, nil
}

func (r *villageRepository) Create(ctx context.Context, village *domain.Village) error {
	// GORM akan otomatis men-generate UUID untuk ID desa dan mengisi CreatedAt/UpdatedAt
	if err := r.db.WithContext(ctx).Create(village).Error; err != nil {
		return err
	}
	return nil
}

// Ambil Daftar Desa dengan Paginasi
func (r *villageRepository) GetPaginated(ctx context.Context, page, limit int) ([]domain.Village, int64, error) {
	var villages []domain.Village
	var totalData int64

	// 1. Base Query
	baseQuery := r.db.WithContext(ctx).Model(&domain.Village{})

	// 2. Hitung Total Data
	if err := baseQuery.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}
	if totalData == 0 {
		return []domain.Village{}, 0, nil
	}

	// 3. Hitung Offset
	offset := (page - 1) * limit

	// 4. Ambil Data (Diurutkan berdasarkan Abjad A-Z)
	err := baseQuery.
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&villages).Error

	if err != nil {
		return nil, 0, err
	}

	return villages, totalData, nil
}

// Fungsi untuk mencari Desa dari titik Koordinat
func (r *villageRepository) GetByCoordinate(ctx context.Context, lat, lon float64) (*domain.Village, error) {
	var village domain.Village

	// PostGIS Query: ST_Intersects (Apakah titik X,Y menyentuh/berada di dalam Polygon boundary desa?)
	// CATATAN PENTING: ST_MakePoint urutannya harus (Longitude/X, Latitude/Y)
	query := "ST_Intersects(boundary, ST_SetSRID(ST_MakePoint(?, ?), 4326))"

	err := r.db.WithContext(ctx).Where(query, lon, lat).First(&village).Error
	if err != nil {
		return nil, err // Akan me-return error (misal: RecordNotFound) jika koordinat di luar area peta
	}

	return &village, nil
}
