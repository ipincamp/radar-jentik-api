package repositories

import (
	"context"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportRepo struct {
	db *gorm.DB
}

type Report struct {
	ID         string      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ReporterID string      `gorm:"type:uuid"`
	VerifierID *string     `gorm:"type:uuid"`
	Location   interface{} `gorm:"type:geometry(Point, 4326)"`
	// Field shadow dengan tag `->` (Read Only)
	// GORM akan mengisi field ini dari hasil query SELECT, tapi mengabaikannya saat INSERT/UPDATE
	LatitudeDB         float64 `gorm:"column:latitude;->"`
	LongitudeDB        float64 `gorm:"column:longitude;->"`
	LarvaeDensityIndex int
	PhotoURL           string
	Notes              string
	Status             string
	VerifiedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt
}

// Update Helper ToDomain untuk menggunakan LatitudeDB/LongitudeDB
func (r *Report) ToDomain() *domain.Report {
	return &domain.Report{
		ID:         r.ID,
		ReporterID: r.ReporterID,
		VerifierID: r.VerifierID,
		// Ambil dari field shadow
		Latitude:           r.LatitudeDB,
		Longitude:          r.LongitudeDB,
		LarvaeDensityIndex: r.LarvaeDensityIndex,
		PhotoURL:           r.PhotoURL,
		Notes:              r.Notes,
		Status:             r.Status,
		VerifiedAt:         r.VerifiedAt,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

func NewReportRepo(db *gorm.DB) ports.ReportRepository {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) Save(ctx context.Context, d *domain.Report) error {
	// Mapping Domain ke DB
	// PENTING: Menggunakan ST_SetSRID(ST_MakePoint(lon, lat), 4326) untuk PostGIS
	reportDB := Report{
		ReporterID:         d.ReporterID,
		LarvaeDensityIndex: d.LarvaeDensityIndex,
		PhotoURL:           d.PhotoURL,
		Notes:              d.Notes,
		Status:             d.Status,
		// Query raw untuk insert geometry
		Location: clause.Expr{
			SQL: "ST_SetSRID(ST_MakePoint(?, ?), 4326)",
			// Perhatikan urutan: Longitude dulu (X), baru Latitude (Y)
			Vars: []interface{}{d.Longitude, d.Latitude},
		},
	}

	if err := r.db.WithContext(ctx).Create(&reportDB).Error; err != nil {
		return err
	}

	d.ID = reportDB.ID
	return nil
}

// Implementasi FindAll
func (r *ReportRepo) FindAll(ctx context.Context, page, limit int) ([]*domain.Report, int64, error) {
	var reports []Report
	var total int64

	offset := (page - 1) * limit

	// 1. Hitung Total Data (untuk pagination)
	if err := r.db.WithContext(ctx).Model(&Report{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. Query Data dengan Ekstraksi PostGIS
	// ST_Y = Latitude, ST_X = Longitude
	query := r.db.WithContext(ctx).
		Select("*, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude").
		Order("created_at DESC"). // Urutkan dari yang terbaru
		Limit(limit).
		Offset(offset).
		Find(&reports)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	// 3. Konversi ke Domain
	domainReports := make([]*domain.Report, len(reports))
	for i, r := range reports {
		domainReports[i] = r.ToDomain()
	}

	return domainReports, total, nil
}

func (r *ReportRepo) FindByID(ctx context.Context, id string) (*domain.Report, error) {
	var reportDB Report
	// Mengambil data termasuk koordinat geometri
	if err := r.db.WithContext(ctx).
		Select("*, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude").
		First(&reportDB, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return reportDB.ToDomain(), nil
}

func (r *ReportRepo) Update(ctx context.Context, d *domain.Report) error {
	// Hanya update field yang relevan untuk proses validasi
	return r.db.WithContext(ctx).Model(&Report{}).
		Where("id = ?", d.ID).
		Updates(map[string]interface{}{
			"status":      d.Status,
			"notes":       d.Notes,
			"verifier_id": d.VerifierID,
			"verified_at": d.VerifiedAt,
			"updated_at":  time.Now(),
		}).Error
}
