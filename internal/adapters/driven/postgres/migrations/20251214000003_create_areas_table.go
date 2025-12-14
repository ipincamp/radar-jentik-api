package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateAreasTable() *gormigrate.Migration {
	type Area struct {
		// ID UUID
		ID string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`

		// Business Fields
		ParentID *string `gorm:"type:uuid;index"` // Nullable untuk root area
		Name     string  `gorm:"type:varchar(100);not null"`
		Type     string  `gorm:"type:varchar(50);not null"` // e.g. "desa", "rw"
		// Menggunakan MultiPolygon agar aman jika area terpisah
		Boundary string `gorm:"type:geometry(MultiPolygon, 4326);not null"`

		// Standard Timestamps (Timestamptz)
		CreatedAt time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}

	return &gormigrate.Migration{
		ID: "20251214000003",
		Migrate: func(tx *gorm.DB) error {
			// 1. Buat Tabel Areas
			if err := tx.AutoMigrate(&Area{}); err != nil {
				return err
			}
			// 2. Tambah kolom area_id ke users (Alter Table)
			return tx.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS area_id uuid REFERENCES areas(id)`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS area_id`).Error; err != nil {
				return err
			}
			return tx.Migrator().DropTable(&Area{})
		},
	}
}
