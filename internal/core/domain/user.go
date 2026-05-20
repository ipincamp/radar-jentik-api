package domain

import "time"

type User struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	FullName  string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Username  string    `gorm:"type:varchar(255);not null;unique" json:"username"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`   // Jangan tampilkan password di JSON
	Role      string    `gorm:"type:varchar(20);not null" json:"role"` // 'cadre' atau 'officer'
	VillageID string    `gorm:"type:uuid;not null" json:"village_id"`  // FK ke Village (UUID)
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relasi
	Village *Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`
}
