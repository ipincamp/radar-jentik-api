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

func (r *malariaSurveyRepository) Update(ctx context.Context, id string, survey *domain.MalariaSurvey) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update data induk menggunakan map agar nilai 0 / boolean false tetap terupdate
		updateData := map[string]interface{}{
			"village_id":          survey.VillageID,
			"dusun":               survey.Dusun,
			"rt":                  survey.RT,
			"rw":                  survey.RW,
			"latitude":            survey.Latitude,
			"longitude":           survey.Longitude,
			"breeding_place_type": survey.BreedingPlaceType,
			"physical_lighting":   survey.PhysicalLighting,
			"physical_water_flow": survey.PhysicalWaterFlow,
			"bio_plants":          survey.BioPlants,
			"bio_animals":         survey.BioAnimals,
			"chem_salinity":       survey.ChemSalinity,
			"chem_ph":             survey.ChemPH,
			"chem_water_temp":     survey.ChemWaterTemp,
			"area":                survey.Area,
			"depth":               survey.Depth,
			"is_larvae_found":     survey.IsLarvaeFound,
			"larvae_species":      survey.LarvaeSpecies,
			"scoop_count":         survey.ScoopCount,
			"larvae_count":        survey.LarvaeCount,
			"larvae_density":      survey.LarvaeDensity,
			"inspected_at":        survey.InspectedAt,
		}

		if err := tx.Model(&domain.MalariaSurvey{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
			return err
		}

		// Update kolom geometri PostGIS
		geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?::float, ?::float), 4326)", survey.Longitude, survey.Latitude)
		if err := tx.Model(&domain.MalariaSurvey{}).Where("id = ?", id).Update("geom", geomExpr).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *malariaSurveyRepository) GetExportData(ctx context.Context, userID, role string) ([]domain.MalariaSurvey, error) {
	var surveys []domain.MalariaSurvey
	query := r.db.WithContext(ctx).Preload("Village")

	// Restriksi jika kader yang mengunduh (hanya desa miliknya)
	if role == "cadre" {
		query = query.Where("village_id = (SELECT village_id FROM users WHERE id = ?)", userID)
	}

	err := query.Joins("JOIN villages ON villages.id = malaria_surveys.village_id").
		Order("villages.name ASC, malaria_surveys.inspected_at DESC").
		Find(&surveys).Error

	return surveys, err
}
