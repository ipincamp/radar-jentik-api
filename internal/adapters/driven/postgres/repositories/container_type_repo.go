package repositories

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"gorm.io/gorm"
)

type containerTypeRepository struct {
	db *gorm.DB
}

func NewContainerTypeRepository(db *gorm.DB) ports.ContainerTypeRepository {
	return &containerTypeRepository{db: db}
}

func (r *containerTypeRepository) FindAllActive(ctx context.Context) ([]domain.ContainerType, error) {
	var types []domain.ContainerType
	// Hanya ambil yang is_active = true dan urutkan sesuai abjad
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("name asc").Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}
