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
