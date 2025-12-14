package domain

import "time"

type Area struct {
	ID        string
	ParentID  *string
	Name      string
	Type      string
	GeoJSON   string `gorm:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
