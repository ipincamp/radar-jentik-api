package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func CreateMalariaSurveysTable() *gormigrate.Migration {
	// Definisikan struct lokal hanya untuk proses auto-migrate
	type MalariaSurvey struct {
		ID                string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
		UserID            string         `gorm:"type:uuid;not null;index"`
		VillageID         string         `gorm:"type:uuid;not null;index"`
		Dusun             string         `gorm:"type:varchar(100)"`
		RT                string         `gorm:"type:varchar(10)"`
		RW                string         `gorm:"type:varchar(10)"`
		Latitude          float64        `gorm:"type:double precision;not null"`
		Longitude         float64        `gorm:"type:double precision;not null"`
		Geom              string         `gorm:"type:geometry(Point,4326)"`
		BreedingPlaceType string         `gorm:"type:varchar(100)"`
		PhysicalLighting  string         `gorm:"type:varchar(100)"`
		PhysicalWaterFlow string         `gorm:"type:varchar(100)"`
		BioPlants         string         `gorm:"type:varchar(255)"`
		BioAnimals        string         `gorm:"type:varchar(255)"`
		ChemSalinity      float64        `gorm:"type:double precision"`
		ChemPH            float64        `gorm:"type:double precision"`
		ChemWaterTemp     float64        `gorm:"type:double precision"`
		Area              float64        `gorm:"type:double precision"`
		Depth             float64        `gorm:"type:double precision"`
		IsLarvaeFound     bool           `gorm:"type:boolean;default:false"`
		LarvaeSpecies     string         `gorm:"type:varchar(100)"`
		ScoopCount        int            `gorm:"type:int;default:0"`
		LarvaeCount       int            `gorm:"type:int;default:0"`
		LarvaeDensity     float64        `gorm:"type:double precision;default:0"`
		InspectedAt       time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		CreatedAt         time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt         time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
		DeletedAt         gorm.DeletedAt `gorm:"type:timestamptz;index"`
	}

	return &gormigrate.Migration{
		ID: "00006",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.AutoMigrate(&MalariaSurvey{}); err != nil {
				return err
			}
			// Index Spasial PostGIS
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_malaria_surveys_geom ON malaria_surveys USING GIST (geom)`).Error; err != nil {
				return err
			}
			// Foreign Keys
			tx.Exec(`ALTER TABLE malaria_surveys ADD CONSTRAINT fk_malaria_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`)
			tx.Exec(`ALTER TABLE malaria_surveys ADD CONSTRAINT fk_malaria_village FOREIGN KEY (village_id) REFERENCES villages(id) ON DELETE RESTRICT`)
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(&MalariaSurvey{})
		},
	}
}
