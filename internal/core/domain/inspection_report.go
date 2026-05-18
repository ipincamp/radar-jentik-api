package domain

import "time"

type InspectionReport struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	VillageID        string    `json:"village_id"`
	RT               string    `json:"rt"`
	RW               string    `json:"rw"`
	FamilyHeadName   string    `json:"family_head_name"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	LarvaeStatus     int       `json:"larvae_status"`
	ValidationStatus string    `json:"validation_status"` // 'pending', 'accept', 'reject'
	InspectedAt      time.Time `json:"inspected_at"`

	// Relasi One-to-Many
	ContainerDetails []ContainerDetail `json:"container_details,omitempty"`
}
