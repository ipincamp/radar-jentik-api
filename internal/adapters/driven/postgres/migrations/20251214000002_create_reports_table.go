package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateReportsTable() *gormigrate.Migration {
	type Report struct {
		ID string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`

		// Foreign Keys
		ReporterID string  `gorm:"type:uuid;not null;index"`
		VerifierID *string `gorm:"type:uuid;index"`

		// Spatial Data (PostGIS)
		Location string `gorm:"type:geometry(Point, 4326);not null"`

		// Business Fields
		LarvaeDensityIndex int    `gorm:"type:int;not null"`
		PhotoURL           string `gorm:"type:text"`
		Notes              string `gorm:"type:text"`
		Status             string `gorm:"type:varchar(50);default:'pending';not null"` // pending, verified, rejected

		// Timestamps
		VerifiedAt *time.Time     `gorm:"type:timestamptz"`
		CreatedAt  time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt  time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "20251214000002",
		Migrate: func(tx *gorm.DB) error {
			// 1. Aktifkan PostGIS extension
			if err := tx.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
				return err
			}
			// 2. Buat Tabel
			return tx.AutoMigrate(&Report{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&Report{})
		},
	}
}
