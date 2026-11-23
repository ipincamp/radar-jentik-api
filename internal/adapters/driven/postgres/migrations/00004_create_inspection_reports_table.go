package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateInspectionReportsTable() *gormigrate.Migration {
	type InspectionReport struct {
		ID               string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		UserID           string    `gorm:"type:uuid;not null;index"`
		VillageID        string    `gorm:"type:uuid;not null;index"`
		RT               string    `gorm:"type:varchar(10);not null"`
		RW               string    `gorm:"type:varchar(10);not null"`
		FamilyHeadName   string    `gorm:"type:varchar(255);not null"`
		Latitude         float64   `gorm:"type:double precision;not null"`
		Longitude        float64   `gorm:"type:double precision;not null"`
		Geom             string    `gorm:"type:geometry(Point,4326)"`
		LarvaeStatus     int       `gorm:"type:smallint;not null"`
		PhotoURL         string    `gorm:"type:varchar(255);not null"`
		ValidationStatus string    `gorm:"type:varchar(20);default:'pending'"`
		RejectionReason  *string   `gorm:"type:text"`
		InspectedAt      time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "00004",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&InspectionReport{}); err != nil {
				return err
			}

			// Index Spasial
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_inspection_reports_geom ON inspection_reports USING GIST (geom)`).Error; err != nil {
				return err
			}

			// Composite Index untuk Peta IDW, Rekapitulasi, & Export (filter validation_status + village_id)
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_inspection_reports_status_village_inspected ON inspection_reports (validation_status, village_id, inspected_at DESC) WHERE deleted_at IS NULL`).Error; err != nil {
				return err
			}

			// Composite Index untuk Riwayat Kader (filter user_id + sorting created_at)
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_inspection_reports_user_created ON inspection_reports (user_id, created_at DESC) WHERE deleted_at IS NULL`).Error; err != nil {
				return err
			}

			// Composite Index untuk Antrean Validasi Petugas
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_inspection_reports_pending_created ON inspection_reports (validation_status, created_at ASC) WHERE deleted_at IS NULL`).Error; err != nil {
				return err
			}

			// Foreign Keys
			tx.Exec(`ALTER TABLE inspection_reports ADD CONSTRAINT fk_report_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`)
			tx.Exec(`ALTER TABLE inspection_reports ADD CONSTRAINT fk_report_village FOREIGN KEY (village_id) REFERENCES villages(id) ON DELETE RESTRICT`)
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&InspectionReport{})
		},
	}
}
