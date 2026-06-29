package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateContainerDetailsTable() *gormigrate.Migration {
	type ContainerDetail struct {
		ID                 string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		InspectionReportID string `gorm:"type:uuid;not null;index"`
		ContainerTypeID    string `gorm:"type:uuid;not null;index"`
		InspectedCount     int    `gorm:"type:int;not null;default:0"`
		PositiveCount      int    `gorm:"type:int;not null;default:0"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "00005",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&ContainerDetail{}); err != nil {
				return err
			}

			// Foreign Keys
			tx.Exec(`ALTER TABLE container_details ADD CONSTRAINT fk_detail_report FOREIGN KEY (inspection_report_id) REFERENCES inspection_reports(id) ON DELETE CASCADE`)
			tx.Exec(`ALTER TABLE container_details ADD CONSTRAINT fk_detail_type FOREIGN KEY (container_type_id) REFERENCES container_types(id) ON DELETE RESTRICT`)

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&ContainerDetail{})
		},
	}
}
