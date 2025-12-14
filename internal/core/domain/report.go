package domain

import "time"

type Report struct {
	ID                 string
	ReporterID         string
	VerifierID         *string
	Latitude           float64 // Representasi dari Location
	Longitude          float64 // Representasi dari Location
	LarvaeDensityIndex int
	PhotoURL           string
	Notes              string
	Status             string
	VerifiedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}
