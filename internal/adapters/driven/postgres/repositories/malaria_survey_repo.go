package repositories

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"gorm.io/gorm"
)

type malariaSurveyRepository struct {
	db *gorm.DB
}

func NewMalariaSurveyRepository(db *gorm.DB) ports.MalariaSurveyRepository {
	return &malariaSurveyRepository{db: db}
}

func (r *malariaSurveyRepository) Create(ctx context.Context, survey *domain.MalariaSurvey) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Omit Geom karena akan di-inject menggunakan query khusus PostGIS
		if err := tx.Omit("Geom").Create(survey).Error; err != nil {
			return err
		}

		// Update kolom 'geom' menggunakan PostGIS
		geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?::float, ?::float), 4326)", survey.Longitude, survey.Latitude)
		if err := tx.Model(survey).Update("geom", geomExpr).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *malariaSurveyRepository) GetPaginatedHistory(ctx context.Context, userID string, page, limit int) ([]domain.MalariaSurvey, int64, error) {
	var surveys []domain.MalariaSurvey
	var totalData int64

	baseQuery := r.db.WithContext(ctx).Model(&domain.MalariaSurvey{}).Where("user_id = ?", userID)
	if err := baseQuery.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := baseQuery.Preload("Village").Order("inspected_at DESC").Limit(limit).Offset(offset).Find(&surveys).Error

	return surveys, totalData, err
}
