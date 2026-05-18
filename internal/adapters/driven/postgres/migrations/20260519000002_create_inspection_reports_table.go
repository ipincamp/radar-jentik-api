package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateInspectionReports() *gormigrate.Migration {
	type InspectionReport struct {
		ID               string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		UserID           string    `gorm:"type:uuid;not null;index"`
		VillageID        string    `gorm:"type:uuid;not null;index"`
		RT               string    `gorm:"type:varchar(10);not null"`
		RW               string    `gorm:"type:varchar(10);not null"`
		FamilyHeadName   string    `gorm:"type:varchar(255)"`
		Latitude         float64   `gorm:"type:numeric(10,8);not null"`
		Longitude        float64   `gorm:"type:numeric(11,8);not null"`
		Geom             string    `gorm:"type:geometry(Point,4326)"`
		LarvaeStatus     int       `gorm:"type:smallint;not null"`
		ValidationStatus string    `gorm:"type:varchar(20);default:'pending'"`
		InspectedAt      time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "20260519000002",

		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
				return err
			}
			// Mengaktifkan fitur geospasial
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS postgis`).Error; err != nil {
				return err
			}

			if err := tx.AutoMigrate(&InspectionReport{}); err != nil {
				return err
			}

			// GIST index secara raw query untuk optimasi performa pencarian spasial
			return tx.Exec(`CREATE INDEX IF NOT EXISTS idx_inspection_reports_geom ON inspection_reports USING GIST (geom)`).Error
		},

		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&InspectionReport{})
		},
	}
}
