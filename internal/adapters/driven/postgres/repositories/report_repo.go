package repositories

import (
	"context"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportRepo struct {
	db *gorm.DB
}

type Report struct {
	ID                 string      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ReporterID         string      `gorm:"type:uuid"`
	VerifierID         *string     `gorm:"type:uuid"`
	Location           interface{} `gorm:"type:geometry(Point, 4326)"`
	LarvaeDensityIndex int
	PhotoURL           string
	Notes              string
	Status             string
	VerifiedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt
}

func NewReportRepo(db *gorm.DB) ports.ReportRepository {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) Save(ctx context.Context, d *domain.Report) error {
	// Mapping Domain ke DB
	// PENTING: Menggunakan ST_SetSRID(ST_MakePoint(lon, lat), 4326) untuk PostGIS
	reportDB := Report{
		ReporterID:         d.ReporterID,
		LarvaeDensityIndex: d.LarvaeDensityIndex,
		PhotoURL:           d.PhotoURL,
		Notes:              d.Notes,
		Status:             d.Status,
		// Query raw untuk insert geometry
		Location: clause.Expr{
			SQL: "ST_SetSRID(ST_MakePoint(?, ?), 4326)",
			// Perhatikan urutan: Longitude dulu (X), baru Latitude (Y)
			Vars: []interface{}{d.Longitude, d.Latitude},
		},
	}

	if err := r.db.WithContext(ctx).Create(&reportDB).Error; err != nil {
		return err
	}

	d.ID = reportDB.ID
	return nil
}
