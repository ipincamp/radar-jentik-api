package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateUsersTable() *gormigrate.Migration {
	type User struct {
		ID        string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		VillageID string `gorm:"type:uuid;not null;index"`
		Name      string `gorm:"type:varchar(100);not null"`
		Username  string `gorm:"type:varchar(100);uniqueIndex;not null"`
		Password  string `gorm:"type:varchar(255);not null"`
		Role      string `gorm:"type:varchar(20);not null"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "00002",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&User{}); err != nil {
				return err
			}
			// Foreign Key
			return tx.Exec(`
				ALTER TABLE users
				ADD CONSTRAINT fk_users_village
				FOREIGN KEY (village_id) REFERENCES villages(id) ON DELETE RESTRICT
			`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&User{})
		},
	}
}
