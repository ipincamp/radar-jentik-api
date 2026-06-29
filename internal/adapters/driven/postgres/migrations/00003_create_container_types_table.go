package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateContainerTypesTable() *gormigrate.Migration {
	type ContainerType struct {
		ID       string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		Name     string `gorm:"type:varchar(100);uniqueIndex;not null"`
		IsActive bool   `gorm:"type:boolean;default:true"`

		CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "00003",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&ContainerType{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&ContainerType{})
		},
	}
}
