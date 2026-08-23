package domain

import (
	"time"

	"gorm.io/gorm"
)

type MalariaSurvey struct {
	ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID    string `gorm:"type:uuid;not null" json:"user_id"`
	VillageID string `gorm:"type:uuid;not null" json:"village_id"`
	Dusun     string `gorm:"type:varchar(100)" json:"dusun"`
	RT        string `gorm:"type:varchar(10)" json:"rt"`
	RW        string `gorm:"type:varchar(10)" json:"rw"`

	// Lokasi & Waktu
	Latitude          float64 `gorm:"type:double precision;not null" json:"latitude"`
	Longitude         float64 `gorm:"type:double precision;not null" json:"longitude"`
	Geom              string  `gorm:"-" json:"-"`                                   // Diisi manual via PostGIS
	BreedingPlaceType string  `gorm:"type:varchar(100)" json:"breeding_place_type"` // sawah/mata air/dll

	// Karakteristik Fisik, Biologi, Kimia
	PhysicalLighting  string  `gorm:"type:varchar(100)" json:"physical_lighting"`   // Naungan/Terkena Sinar Matahari
	PhysicalWaterFlow string  `gorm:"type:varchar(100)" json:"physical_water_flow"` // Aliran Air
	BioPlants         string  `gorm:"type:varchar(255)" json:"bio_plants"`          // Tanaman Air/Alga/Lumut
	BioAnimals        string  `gorm:"type:varchar(255)" json:"bio_animals"`         // Hewan Lain
	ChemSalinity      float64 `gorm:"type:double precision" json:"chem_salinity"`   // Kadar Garam
	ChemPH            float64 `gorm:"type:double precision" json:"chem_ph"`
	ChemWaterTemp     float64 `gorm:"type:double precision" json:"chem_water_temp"`
	Area              float64 `gorm:"type:double precision" json:"area"`  // Luas Tempat Perindukan
	Depth             float64 `gorm:"type:double precision" json:"depth"` // Kedalaman

	// Data Jentik
	IsLarvaeFound bool    `gorm:"type:boolean;default:false" json:"is_larvae_found"`
	LarvaeSpecies string  `gorm:"type:varchar(100)" json:"larvae_species"`               // Anopheles/Aedes/Culex
	ScoopCount    int     `gorm:"type:int;default:0" json:"scoop_count"`                 // Jumlah Cidukan
	LarvaeCount   int     `gorm:"type:int;default:0" json:"larvae_count"`                // Jumlah Larva
	LarvaeDensity float64 `gorm:"type:double precision;default:0" json:"larvae_density"` // Kepadatan

	InspectedAt time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"inspected_at"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relasi
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Village *Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`
}
