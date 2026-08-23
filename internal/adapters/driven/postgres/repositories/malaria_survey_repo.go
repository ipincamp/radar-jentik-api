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
