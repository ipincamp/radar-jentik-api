package domain

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Village struct {
	ID   string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	// Menyimpan data polygon/multipolygon dalam format GeoJSON
	Boundary json.RawMessage `gorm:"type:jsonb" json:"boundary"`

	CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete
}
