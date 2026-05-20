package repositories

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"

	"gorm.io/gorm"
)

type inspectionReportRepository struct {
	db *gorm.DB
}

func NewInspectionReportRepository(db *gorm.DB) ports.InspectionReportRepository {
	return &inspectionReportRepository{db: db}
}

func (r *inspectionReportRepository) Create(ctx context.Context, report *domain.InspectionReport) error {
	// Memulai Transaksi Database
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1. GORM secara otomatis akan menyimpan 'report' (induk)
		//    DAN data 'ContainerDetails' (anak) jika relasinya sudah di-set di struct.
		//    Namun, kita mengecualikan kolom Geom karena butuh fungsi PostGIS khusus.
		if err := tx.Omit("Geom").Create(report).Error; err != nil {
			// Jika error, transaksi otomatis di-rollback
			return err
		}

		// 2. Update kolom 'geom' menggunakan fungsi spasial PostGIS ST_SetSRID dan ST_MakePoint.
		//    Bujur (Longitude) adalah X, Lintang (Latitude) adalah Y.
		geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?::float, ?::float), 4326)", report.Longitude, report.Latitude)

		if err := tx.Model(report).Update("geom", geomExpr).Error; err != nil {
			// Jika gagal update spasial, transaksi otomatis di-rollback
			return err
		}

		// Jika sampai di sini, kembalikan nil untuk COMMIT transaksi
		return nil
	})

	return err
}

// Ambil riwayat laporan milik kader tertentu
func (r *inspectionReportRepository) GetByUserID(ctx context.Context, userID string) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport
	// Preload digunakan untuk mengambil data wadah (anak) sekaligus
	err := r.db.WithContext(ctx).Preload("ContainerDetails").Where("user_id = ?", userID).Order("inspected_at desc").Find(&reports).Error
	return reports, err
}

// Ambil semua laporan yang menunggu validasi petugas
func (r *inspectionReportRepository) GetPending(ctx context.Context) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport
	err := r.db.WithContext(ctx).Preload("ContainerDetails").Where("validation_status = ?", "pending").Order("inspected_at asc").Find(&reports).Error
	return reports, err
}

// Ubah status laporan (Terima/Tolak)
func (r *inspectionReportRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&domain.InspectionReport{}).Where("id = ?", id).Update("validation_status", status).Error
}

// Ambil data untuk Peta Zonasi IDW (Hanya yang berstatus 'accept')
func (r *inspectionReportRepository) GetValidReports(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport

	// Query dasar: Hanya ambil laporan yang sudah divalidasi (accept)
	query := r.db.WithContext(ctx).Where("validation_status = ?", "accept")

	// RESTRIKSI LINGKUP (ISOLASI DATA)
	if role == "cadre" {
		// Jika dia kader, filter laporan yang village_id-nya sama dengan village_id milik kader tersebut
		query = query.Where("village_id = (SELECT village_id FROM users WHERE id = ?)", userID)
	}
	// Jika role == "officer" (Petugas Puskesmas), kita tidak menambahkan Where village_id,
	// sehingga petugas bisa melihat SEMUA titik laporan di semua desa.

	err := query.Order("inspected_at desc").Find(&reports).Error

	return reports, err
}
