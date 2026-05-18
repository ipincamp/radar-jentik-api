package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateContainerDetails() *gormigrate.Migration {
	type ContainerDetail struct {
		ID                 string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		InspectionReportID string `gorm:"type:uuid;not null;index"`
		ContainerType      string `gorm:"type:varchar(50);not null"`
		InspectedCount     int    `gorm:"type:int;not null;default:0"`
		PositiveCount      int    `gorm:"type:int;not null;default:0"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "20260519000003",

		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
				return err
			}

			if err := tx.AutoMigrate(&ContainerDetail{}); err != nil {
				return err
			}

			// Foreign Key di level database
			return tx.Exec(`
				ALTER TABLE container_details 
				ADD CONSTRAINT fk_inspection_report 
				FOREIGN KEY (inspection_report_id) 
				REFERENCES inspection_reports(id) 
				ON DELETE CASCADE
			`).Error
		},

		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&ContainerDetail{})
		},
	}
}
