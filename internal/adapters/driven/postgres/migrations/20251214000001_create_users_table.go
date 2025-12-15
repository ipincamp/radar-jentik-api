package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateUsersTable() *gormigrate.Migration {
	type User struct {
		// ID UUID
		ID string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`

		// Business Fields
		Name     string `gorm:"type:varchar(100)"`
		Username string `gorm:"type:varchar(100);uniqueIndex;not null"`
		Password string `gorm:"type:varchar(255);not null"`
		Role     string `gorm:"type:varchar(20);not null"`

		// Standard Timestamps (Timestamptz)
		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "20251214000001",

		Migrate: func(tx *gorm.DB) error {
			// Extension UUID
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
				return err
			}

			return tx.AutoMigrate(&User{})
		},

		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&User{})
		},
	}
}
