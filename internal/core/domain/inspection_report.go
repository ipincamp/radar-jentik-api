package domain

import (
	"time"

	"gorm.io/gorm"
)

type InspectionReport struct {
	ID             string  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID         string  `gorm:"type:uuid;not null" json:"user_id"`
	VillageID      string  `gorm:"type:uuid;not null" json:"village_id"`
	RT             string  `gorm:"type:varchar(10);not null" json:"rt"`
	RW             string  `gorm:"type:varchar(10);not null" json:"rw"`
	FamilyHeadName string  `gorm:"type:varchar(255)" json:"family_head_name"`
	Latitude       float64 `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude      float64 `gorm:"type:decimal(11,8);not null" json:"longitude"`

	// Tipe Spasial PostGIS untuk IDW
	Geom string `gorm:"type:geometry(Point,4326)" json:"-"`

	LarvaeStatus     int       `gorm:"type:int;not null" json:"larvae_status"`                      // 0: negative, 1: positive
	ValidationStatus string    `gorm:"type:varchar(20);default:'pending'" json:"validation_status"` // pending, accept, reject
	InspectedAt      time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"inspected_at"`

	CreatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relasi
	User             *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Village          *Village          `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	ContainerDetails []ContainerDetail `gorm:"foreignKey:InspectionReportID" json:"container_details,omitempty"`
}
