package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateVillagesTable() *gormigrate.Migration {
	type Village struct {
		ID       string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		Name     string `gorm:"type:varchar(100);not null"`
		Boundary string `gorm:"type:geometry(MultiPolygon, 4326)"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "00001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS postgis`).Error; err != nil {
				return err
			}
			return tx.AutoMigrate(&Village{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&Village{})
		},
	}
}
