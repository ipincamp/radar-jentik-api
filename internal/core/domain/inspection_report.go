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

	// Gunakan gorm:"-" agar GORM tidak mencoba melakukan SELECT/SCAN pada kolom biner geom ini
	Geom string `gorm:"-" json:"-"`

	LarvaeStatus     int            `gorm:"type:int;not null" json:"larvae_status"`
	ValidationStatus string         `gorm:"type:varchar(20);default:'pending'" json:"validation_status"`
	InspectedAt      time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"inspected_at"`
	RejectionReason  *string        `json:"rejection_reason" db:"rejection_reason"`
	CreatedAt        time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relasi
	User             *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Village          *Village          `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	ContainerDetails []ContainerDetail `gorm:"foreignKey:InspectionReportID" json:"container_details,omitempty"`
}

type ReportRecap struct {
	RT                string
	RumahDiperiksa    int
	RumahPositif      int
	BakMandiTotal     int
	BakMandiPos       int
	TempayanTotal     int
	TempayanPos       int
	PecahanBotolTotal int
	PecahanBotolPos   int
	BarangBekasTotal  int
	BarangBekasPos    int
	KulkasTotal       int
	KulkasPos         int
	TandonAirTotal    int
	TandonAirPos      int
	VasBungaTotal     int
	VasBungaPos       int
	PotBungaTotal     int
	PotBungaPos       int
	LainLainTotal     int
	LainLainPos       int
	TotalContainer    int
	TotalContainerPos int
}

type ValidateReportRequest struct {
	Status          string `json:"status"`           // Isinya "accept" atau "reject"
	RejectionReason string `json:"rejection_reason"` // Diisi alasan jika status "reject"
}
