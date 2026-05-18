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
		geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", report.Longitude, report.Latitude)

		if err := tx.Model(report).UpdateColumn("geom", geomExpr).Error; err != nil {
			// Jika gagal update spasial, transaksi otomatis di-rollback
			return err
		}

		// Jika sampai di sini, kembalikan nil untuk COMMIT transaksi
		return nil
	})

	return err
}
