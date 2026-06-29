package domain

import (
	"time"

	"gorm.io/gorm"
)

type InspectionReport struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID           string         `gorm:"type:uuid;not null" json:"user_id"`
	VillageID        string         `gorm:"type:uuid;not null" json:"village_id"`
	RT               string         `gorm:"type:varchar(10);not null" json:"rt"`
	RW               string         `gorm:"type:varchar(10);not null" json:"rw"`
	FamilyHeadName   string         `gorm:"type:varchar(255);not null" json:"family_head_name"`
	Latitude         float64        `gorm:"type:double precision;not null" json:"latitude"`  // Selaras dengan float64 Go & Postgres
	Longitude        float64        `gorm:"type:double precision;not null" json:"longitude"` // Selaras dengan float64 Go & Postgres
	Geom             string         `gorm:"-" json:"-"`                                      // Di-omit karena diisi manual via PostGIS di repo
	LarvaeStatus     int            `gorm:"type:smallint;not null" json:"larvae_status"`     // 0 = Aman, 1 = Positif Jentik
	PhotoURL         string         `gorm:"type:varchar(255);not null" json:"photo_url"`     // Bukti foto lapangan wajib
	ValidationStatus string         `gorm:"type:varchar(20);default:'pending'" json:"validation_status"`
	RejectionReason  *string        `gorm:"type:text" json:"rejection_reason"` // Menggunakan tipe pointer agar bisa bernilai NULL di database
	InspectedAt      time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"inspected_at"`
	CreatedAt        time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relasi Model
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
	Status          string `json:"status"`           // "accept" atau "reject"
	RejectionReason string `json:"rejection_reason"` // Keterangan jika ditolak petugas
}
