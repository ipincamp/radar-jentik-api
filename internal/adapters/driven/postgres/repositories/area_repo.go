package repositories

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AreaRepo struct {
	db *gorm.DB
}

func NewAreaRepo(db *gorm.DB) ports.AreaRepository {
	return &AreaRepo{db: db}
}

type Area struct {
	ID       string      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name     string      `gorm:"type:varchar(100)"`
	Type     string      `gorm:"type:varchar(50)"`
	Boundary interface{} `gorm:"type:geometry(MultiPolygon, 4326)"` // Untuk clause.Expr
}

func (r *AreaRepo) Save(ctx context.Context, area *domain.Area, geoJSONGeometry string) error {
	// Query raw untuk insert geometry dari GeoJSON
	areaDB := Area{
		Name: area.Name,
		Type: area.Type,
		Boundary: clause.Expr{
			SQL:  "ST_SetSRID(ST_GeomFromGeoJSON(?), 4326)",
			Vars: []interface{}{geoJSONGeometry},
		},
	}

	// ID harus di-generate oleh DB (karena tag default) dan Boundary adalah kolom geometri.
	if err := r.db.WithContext(ctx).Table("areas").Create(&areaDB).Error; err != nil {
		return err
	}

	// Kembalikan ID yang baru digenerate ke domain
	area.ID = areaDB.ID

	return nil
}

func (r *AreaRepo) FindAll(ctx context.Context) ([]*domain.Area, error) {
	var results []struct {
		ID      string
		Name    string
		Type    string
		GeoJSON string `gorm:"column:geojson"` // Hasil ST_AsGeoJSON
	}

	// Select dengan konversi Geometry ke GeoJSON string
	err := r.db.WithContext(ctx).Table("areas").
		Select("id, name, type, ST_AsGeoJSON(boundary) as geojson").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Mapping ke Domain
	areas := make([]*domain.Area, len(results))
	for i, v := range results {
		areas[i] = &domain.Area{
			ID:      v.ID,
			Name:    v.Name,
			Type:    v.Type,
			GeoJSON: v.GeoJSON, // Ini string JSON mentah
		}
	}
	return areas, nil
}
