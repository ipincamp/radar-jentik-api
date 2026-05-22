package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateVillages() *gormigrate.Migration {
	type Village struct {
		ID       string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		Name     string `gorm:"type:varchar(100);not null"`
		Boundary []byte `gorm:"type:jsonb"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "20260519000001",

		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
				return err
			}
			return tx.AutoMigrate(&Village{})
		},

		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&Village{})
		},
	}
}
