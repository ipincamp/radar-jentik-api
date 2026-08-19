package repositories

import (
	"context"
	"errors"

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
func (r *inspectionReportRepository) UpdateStatus(ctx context.Context, reportID string, status string, rejectionReason *string) error {
	// Query UPDATE sekarang menyertakan rejection_reason
	query := `
		UPDATE inspection_reports 
		SET validation_status = $1, rejection_reason = $2, updated_at = NOW() 
		WHERE id = $3
	`

	tx := r.db.WithContext(ctx).Exec(query, status, rejectionReason, reportID)
	return tx.Error
}

// Ambil data untuk Peta Zonasi IDW (Hanya yang berstatus 'accept')
func (r *inspectionReportRepository) GetValidReports(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport

	query := r.db.WithContext(ctx).Preload("Village").Where("validation_status = ?", "accept")

	// RESTRIKSI LINGKUP (ISOLASI DATA)
	if role == "cadre" {
		// Jika dia kader, filter laporan yang village_id-nya sama dengan village_id milik kader tersebut
		query = query.Where("village_id = (SELECT village_id FROM users WHERE id = ?)", userID)
	}

	err := query.Order("inspected_at desc").Find(&reports).Error
	return reports, err
}

func (r *inspectionReportRepository) GetRecapData(ctx context.Context, userID, role string) ([]domain.ReportRecap, error) {
	// PENTING: Perhatikan penambahan "LEFT JOIN container_types ct"
	// dan perubahan cd.container_type menjadi ct.name
	query := `
		SELECT
			ir.rt,
			COUNT(DISTINCT ir.id) as rumah_diperiksa,
			COUNT(DISTINCT CASE WHEN ir.larvae_status = 1 THEN ir.id END) as rumah_positif,
			COALESCE(SUM(CASE WHEN ct.name = 'Bak Kamar Mandi' THEN cd.inspected_count ELSE 0 END), 0) as bak_mandi_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Bak Kamar Mandi' THEN cd.positive_count ELSE 0 END), 0) as bak_mandi_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Tempayan' THEN cd.inspected_count ELSE 0 END), 0) as tempayan_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Tempayan' THEN cd.positive_count ELSE 0 END), 0) as tempayan_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Pecahan Botol/Air Kemasan' THEN cd.inspected_count ELSE 0 END), 0) as pecahan_botol_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Pecahan Botol/Air Kemasan' THEN cd.positive_count ELSE 0 END), 0) as pecahan_botol_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Barang Bekas' THEN cd.inspected_count ELSE 0 END), 0) as barang_bekas_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Barang Bekas' THEN cd.positive_count ELSE 0 END), 0) as barang_bekas_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Kulkas/Dispenser' THEN cd.inspected_count ELSE 0 END), 0) as kulkas_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Kulkas/Dispenser' THEN cd.positive_count ELSE 0 END), 0) as kulkas_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Tandon Air' THEN cd.inspected_count ELSE 0 END), 0) as tandon_air_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Tandon Air' THEN cd.positive_count ELSE 0 END), 0) as tandon_air_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Vas Bunga' THEN cd.inspected_count ELSE 0 END), 0) as vas_bunga_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Vas Bunga' THEN cd.positive_count ELSE 0 END), 0) as vas_bunga_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Pot Bunga' THEN cd.inspected_count ELSE 0 END), 0) as pot_bunga_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Pot Bunga' THEN cd.positive_count ELSE 0 END), 0) as pot_bunga_pos,
			COALESCE(SUM(CASE WHEN ct.name = 'Lain-lain' THEN cd.inspected_count ELSE 0 END), 0) as lain_lain_total,
			COALESCE(SUM(CASE WHEN ct.name = 'Lain-lain' THEN cd.positive_count ELSE 0 END), 0) as lain_lain_pos,
			COALESCE(SUM(cd.inspected_count), 0) as total_container,
			COALESCE(SUM(cd.positive_count), 0) as total_container_pos
		FROM inspection_reports ir
		LEFT JOIN container_details cd ON ir.id = cd.inspection_report_id
		LEFT JOIN container_types ct ON cd.container_type_id = ct.id
		WHERE ir.validation_status = 'accept'
	`
	// Jika role adalah kader, batasi hanya melihat data miliknya
	var args []interface{}
	if role == "cadre" {
		query += " AND ir.user_id = $1"
		args = append(args, userID)
	}

	// Grouping berdasarkan RT dan urutkan
	query += " GROUP BY ir.rt ORDER BY ir.rt ASC"

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recaps []domain.ReportRecap
	for rows.Next() {
		var rec domain.ReportRecap
		err := rows.Scan(
			&rec.RT, &rec.RumahDiperiksa, &rec.RumahPositif,
			&rec.BakMandiTotal, &rec.BakMandiPos, &rec.TempayanTotal, &rec.TempayanPos,
			&rec.PecahanBotolTotal, &rec.PecahanBotolPos, &rec.BarangBekasTotal, &rec.BarangBekasPos,
			&rec.KulkasTotal, &rec.KulkasPos, &rec.TandonAirTotal, &rec.TandonAirPos,
			&rec.VasBungaTotal, &rec.VasBungaPos, &rec.PotBungaTotal, &rec.PotBungaPos,
			&rec.LainLainTotal, &rec.LainLainPos, &rec.TotalContainer, &rec.TotalContainerPos,
		)
		if err != nil {
			return nil, err
		}
		recaps = append(recaps, rec)
	}

	return recaps, nil
}

func (r *inspectionReportRepository) GetExportData(ctx context.Context, userID, role string) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport

	// Kita ambil laporan beserta Relasi Desa dan Relasi Jenis Wadah-nya
	query := r.db.WithContext(ctx).
		Preload("Village").
		Preload("ContainerDetails.ContainerType").
		Where("validation_status = ?", "accept")

	if role == "cadre" {
		query = query.Where("village_id = (SELECT village_id FROM users WHERE id = ?)", userID)
	}

	// Lakukan JOIN agar bisa mengurutkan dari A-Z berdasarkan Nama Desa, lalu RW, lalu RT
	err := query.Joins("JOIN villages ON villages.id = inspection_reports.village_id").
		Order("villages.name ASC, inspection_reports.rw ASC, inspection_reports.rt ASC").
		Find(&reports).Error

	return reports, err
}

// Ambil Riwayat Laporan Milik 1 Kader (Paginated dengan Filter)
func (r *inspectionReportRepository) GetPaginatedHistory(ctx context.Context, userID string, page, limit int, search, rt, rw, villageID, date string) ([]domain.InspectionReport, int64, error) {
	var reports []domain.InspectionReport
	var totalData int64

	// Base Query: Hanya laporan milik kader ini
	baseQuery := r.db.WithContext(ctx).Model(&domain.InspectionReport{}).Where("user_id = ?", userID)

	// FILTER
	if search != "" {
		// Menggunakan ILIKE untuk case-insensitive di PostgreSQL
		baseQuery = baseQuery.Where("family_head_name ILIKE ?", "%"+search+"%")
	}
	if rt != "" {
		baseQuery = baseQuery.Where("rt = ?", rt)
	}
	if rw != "" {
		baseQuery = baseQuery.Where("rw = ?", rw)
	}
	if villageID != "" {
		baseQuery = baseQuery.Where("village_id = ?", villageID)
	}
	if date != "" {
		// Mengambil laporan berdasarkan tanggal (mengabaikan jam)
		baseQuery = baseQuery.Where("DATE(inspected_at) = ?", date)
	}

	// Hitung total data (setelah filter diterapkan)
	if err := baseQuery.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	if totalData == 0 {
		return []domain.InspectionReport{}, 0, nil
	}

	offset := (page - 1) * limit

	// Ambil data
	err := baseQuery.
		Preload("Village").
		Preload("ContainerDetails.ContainerType").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	if err != nil {
		return nil, 0, err
	}

	return reports, totalData, nil
}

// Ambil Daftar Antrean Laporan untuk Petugas (Paginated)
func (r *inspectionReportRepository) GetPaginatedPending(ctx context.Context, page, limit int) ([]domain.InspectionReport, int64, error) {
	var reports []domain.InspectionReport
	var totalData int64

	// Base Query: Hanya status 'pending'
	baseQuery := r.db.WithContext(ctx).Model(&domain.InspectionReport{}).Where("validation_status = ?", "pending")

	if err := baseQuery.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}
	if totalData == 0 {
		return []domain.InspectionReport{}, 0, nil
	}

	offset := (page - 1) * limit

	// Ambil data (Urutkan dari yang paling LAMA / ASC, seperti antrean loket kasir / FIFO)
	err := baseQuery.
		Preload("User"). // Tampilkan siapa kadernya
		Preload("Village").
		Preload("ContainerDetails.ContainerType").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	if err != nil {
		return nil, 0, err
	}

	return reports, totalData, nil
}

// Fungsi untuk menyimpan banyak laporan sekaligus (bulk insert)
func (r *inspectionReportRepository) CreateBulk(ctx context.Context, reports []*domain.InspectionReport) error {
	// Gunakan transaksi agar jika 1 gagal, semua dibatalkan (konsisten)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, report := range reports {
			// 1. Simpan Laporan Induk dan Detail Wadahnya
			if err := tx.Omit("Geom").Create(report).Error; err != nil {
				return err
			}

			// 2. Buat Spasial Titik PostGIS untuk laporan ini
			geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?::float, ?::float), 4326)", report.Longitude, report.Latitude)

			// Update kolom geom
			if err := tx.Model(report).Update("geom", geomExpr).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkValidateReports bertugas memvalidasi banyak laporan sekaligus (accept/reject) berdasarkan ID yang diberikan.
func (r *inspectionReportRepository) BulkValidateReports(ctx context.Context, reportIDs []string, status string, rejectionReason *string) error {
	// Validasi status yang diterima
	if status != "accept" && status != "reject" {
		return errors.New("status validasi tidak dikenali")
	}

	// Persiapkan data untuk di-update
	updateData := map[string]interface{}{
		"validation_status": status,
		"rejection_reason":  rejectionReason,
	}

	// Update banyak laporan sekaligus menggunakan WHERE id IN (...)
	result := r.db.WithContext(ctx).Model(&domain.InspectionReport{}).
		Where("id IN ?", reportIDs).
		Updates(updateData) // Ganti Update() menjadi Updates()

	return result.Error
}

func (r *inspectionReportRepository) Update(ctx context.Context, reportID string, report *domain.InspectionReport) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Hapus detail wadah lama agar tidak menumpuk
		if err := tx.Where("inspection_report_id = ?", reportID).Delete(&domain.ContainerDetail{}).Error; err != nil {
			return err
		}

		// 2. Update data induk (gunakan map agar GORM memperbarui nilai 0 atau string kosong)
		updateData := map[string]interface{}{
			"village_id":        report.VillageID,
			"rt":                report.RT,
			"rw":                report.RW,
			"family_head_name":  report.FamilyHeadName,
			"latitude":          report.Latitude,
			"longitude":         report.Longitude,
			"larvae_status":     report.LarvaeStatus,
			"photo_url":         report.PhotoURL,
			"inspected_at":      report.InspectedAt,
			"validation_status": "pending", // Reset status menjadi pending setiap kali diedit
		}

		if err := tx.Model(&domain.InspectionReport{}).Where("id = ?", reportID).Updates(updateData).Error; err != nil {
			return err
		}

		// 3. Update spatial point PostGIS
		geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?::float, ?::float), 4326)", report.Longitude, report.Latitude)
		if err := tx.Model(&domain.InspectionReport{}).Where("id = ?", reportID).Update("geom", geomExpr).Error; err != nil {
			return err
		}

		// 4. Masukkan wadah baru yang diedit
		for i := range report.ContainerDetails {
			report.ContainerDetails[i].InspectionReportID = reportID
		}
		if len(report.ContainerDetails) > 0 {
			if err := tx.Create(&report.ContainerDetails).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
