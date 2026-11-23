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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Geom").Create(report).Error; err != nil {
			return err
		}

		geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?::float, ?::float), 4326)", report.Longitude, report.Latitude)
		if err := tx.Model(report).Update("geom", geomExpr).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *inspectionReportRepository) GetByUserID(ctx context.Context, userID string) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport
	err := r.db.WithContext(ctx).
		Preload("ContainerDetails").
		Where("user_id = ?", userID).
		Order("inspected_at DESC").
		Limit(500).
		Find(&reports).Error
	return reports, err
}

func (r *inspectionReportRepository) GetPending(ctx context.Context) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport
	err := r.db.WithContext(ctx).
		Preload("ContainerDetails").
		Where("validation_status = ?", "pending").
		Order("inspected_at ASC").
		Limit(500).
		Find(&reports).Error
	return reports, err
}

func (r *inspectionReportRepository) UpdateStatus(ctx context.Context, reportID string, status string, rejectionReason *string) error {
	query := `
		UPDATE inspection_reports 
		SET validation_status = $1, rejection_reason = $2, updated_at = NOW() 
		WHERE id = $3
	`
	return r.db.WithContext(ctx).Exec(query, status, rejectionReason, reportID).Error
}

// Optimasi: Select kolom esensial saja (tanpa photo_url/metadata berat) & single JOIN Village
func (r *inspectionReportRepository) GetValidReports(ctx context.Context, userID string, role string) ([]domain.InspectionReport, error) {
	var reports []domain.InspectionReport

	query := r.db.WithContext(ctx).
		Select(
			"inspection_reports.id",
			"inspection_reports.village_id",
			"inspection_reports.rt",
			"inspection_reports.rw",
			"inspection_reports.family_head_name",
			"inspection_reports.latitude",
			"inspection_reports.longitude",
			"inspection_reports.larvae_status",
			"inspection_reports.inspected_at",
		).
		Joins("Village").
		Where("inspection_reports.validation_status = ?", "accept")

	if role == "cadre" {
		query = query.Where("inspection_reports.village_id = (SELECT village_id FROM users WHERE id = ?)", userID)
	}

	err := query.Order("inspection_reports.inspected_at DESC").Find(&reports).Error
	return reports, err
}

// Optimasi: Pisahkan agregasi rumah dan wadah agar tidak terjadi Cartesian product saat JOIN
func (r *inspectionReportRepository) GetRecapData(ctx context.Context, userID, role string) ([]domain.ReportRecap, error) {
	var filterClause string
	var args []interface{}

	if role == "cadre" {
		filterClause = " AND ir.user_id = $1"
		args = append(args, userID)
	}

	query := `
		WITH filtered_reports AS (
			SELECT id, rt, larvae_status
			FROM inspection_reports ir
			WHERE ir.validation_status = 'accept' AND ir.deleted_at IS NULL` + filterClause + `
		),
		house_stats AS (
			SELECT
				rt,
				COUNT(*) AS rumah_diperiksa,
				COUNT(*) FILTER (WHERE larvae_status = 1) AS rumah_positif
			FROM filtered_reports
			GROUP BY rt
		),
		container_stats AS (
			SELECT
				fr.rt,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Bak Kamar Mandi'), 0) AS bak_mandi_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Bak Kamar Mandi'), 0) AS bak_mandi_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Tempayan'), 0) AS tempayan_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Tempayan'), 0) AS tempayan_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Pecahan Botol/Air Kemasan'), 0) AS pecahan_botol_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Pecahan Botol/Air Kemasan'), 0) AS pecahan_botol_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Barang Bekas'), 0) AS barang_bekas_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Barang Bekas'), 0) AS barang_bekas_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Kulkas/Dispenser'), 0) AS kulkas_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Kulkas/Dispenser'), 0) AS kulkas_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Tandon Air'), 0) AS tandon_air_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Tandon Air'), 0) AS tandon_air_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Vas Bunga'), 0) AS vas_bunga_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Vas Bunga'), 0) AS vas_bunga_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Pot Bunga'), 0) AS pot_bunga_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Pot Bunga'), 0) AS pot_bunga_pos,
				COALESCE(SUM(cd.inspected_count) FILTER (WHERE ct.name = 'Lain-lain'), 0) AS lain_lain_total,
				COALESCE(SUM(cd.positive_count) FILTER (WHERE ct.name = 'Lain-lain'), 0) AS lain_lain_pos,
				COALESCE(SUM(cd.inspected_count), 0) AS total_container,
				COALESCE(SUM(cd.positive_count), 0) AS total_container_pos
			FROM filtered_reports fr
			LEFT JOIN container_details cd ON fr.id = cd.inspection_report_id
			LEFT JOIN container_types ct ON cd.container_type_id = ct.id
			GROUP BY fr.rt
		)
		SELECT
			hs.rt, hs.rumah_diperiksa, hs.rumah_positif,
			COALESCE(cs.bak_mandi_total, 0), COALESCE(cs.bak_mandi_pos, 0),
			COALESCE(cs.tempayan_total, 0), COALESCE(cs.tempayan_pos, 0),
			COALESCE(cs.pecahan_botol_total, 0), COALESCE(cs.pecahan_botol_pos, 0),
			COALESCE(cs.barang_bekas_total, 0), COALESCE(cs.barang_bekas_pos, 0),
			COALESCE(cs.kulkas_total, 0), COALESCE(cs.kulkas_pos, 0),
			COALESCE(cs.tandon_air_total, 0), COALESCE(cs.tandon_air_pos, 0),
			COALESCE(cs.vas_bunga_total, 0), COALESCE(cs.vas_bunga_pos, 0),
			COALESCE(cs.pot_bunga_total, 0), COALESCE(cs.pot_bunga_pos, 0),
			COALESCE(cs.lain_lain_total, 0), COALESCE(cs.lain_lain_pos, 0),
			COALESCE(cs.total_container, 0), COALESCE(cs.total_container_pos, 0)
		FROM house_stats hs
		LEFT JOIN container_stats cs ON hs.rt = cs.rt
		ORDER BY hs.rt ASC
	`

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

	query := r.db.WithContext(ctx).
		Joins("Village").
		Preload("ContainerDetails.ContainerType").
		Where("inspection_reports.validation_status = ?", "accept")

	if role == "cadre" {
		query = query.Where("inspection_reports.village_id = (SELECT village_id FROM users WHERE id = ?)", userID)
	}

	err := query.
		Order(`"Village".name ASC, inspection_reports.rw ASC, inspection_reports.rt ASC`).
		Find(&reports).Error

	return reports, err
}

func (r *inspectionReportRepository) GetPaginatedHistory(ctx context.Context, userID string, page, limit int, search, rt, rw, villageID, date string) ([]domain.InspectionReport, int64, error) {
	var reports []domain.InspectionReport
	var totalData int64

	baseQuery := r.db.WithContext(ctx).Model(&domain.InspectionReport{}).Where("inspection_reports.user_id = ?", userID)

	if search != "" {
		baseQuery = baseQuery.Where("inspection_reports.family_head_name ILIKE ?", "%"+search+"%")
	}
	if rt != "" {
		baseQuery = baseQuery.Where("inspection_reports.rt = ?", rt)
	}
	if rw != "" {
		baseQuery = baseQuery.Where("inspection_reports.rw = ?", rw)
	}
	if villageID != "" {
		baseQuery = baseQuery.Where("inspection_reports.village_id = ?", villageID)
	}
	if date != "" {
		// Gunakan range waktu agar index pada inspected_at tetap terpakai (DATE(inspected_at) mematikan index!)
		baseQuery = baseQuery.Where("inspection_reports.inspected_at >= ?::date AND inspection_reports.inspected_at < (?::date + INTERVAL '1 day')", date, date)
	}

	if err := baseQuery.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	if totalData == 0 {
		return []domain.InspectionReport{}, 0, nil
	}

	offset := (page - 1) * limit

	err := baseQuery.
		Joins("Village").
		Preload("ContainerDetails.ContainerType").
		Order("inspection_reports.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	if err != nil {
		return nil, 0, err
	}

	return reports, totalData, nil
}

func (r *inspectionReportRepository) GetPaginatedPending(ctx context.Context, page, limit int) ([]domain.InspectionReport, int64, error) {
	var reports []domain.InspectionReport
	var totalData int64

	baseQuery := r.db.WithContext(ctx).Model(&domain.InspectionReport{}).Where("inspection_reports.validation_status = ?", "pending")

	if err := baseQuery.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}
	if totalData == 0 {
		return []domain.InspectionReport{}, 0, nil
	}

	offset := (page - 1) * limit

	err := baseQuery.
		Joins("User").
		Joins("Village").
		Preload("ContainerDetails.ContainerType").
		Order("inspection_reports.created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error

	if err != nil {
		return nil, 0, err
	}

	return reports, totalData, nil
}

// Optimasi: Batch Insert + Single Bulk Spatial Update (dari 2N query menjadi 2 query)
func (r *inspectionReportRepository) CreateBulk(ctx context.Context, reports []*domain.InspectionReport) error {
	if len(reports) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Insert batch sekaligus (ukuran batch 100 agar aman dari batas parameter PostgreSQL)
		if err := tx.Omit("Geom").CreateInBatches(reports, 100).Error; err != nil {
			return err
		}

		// 2. Update seluruh kolom geom sekaligus dalam 1 query SQL
		ids := make([]string, 0, len(reports))
		for _, rep := range reports {
			ids = append(ids, rep.ID)
		}

		if err := tx.Exec(`
			UPDATE inspection_reports
			SET geom = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)
			WHERE id IN ?
		`, ids).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *inspectionReportRepository) BulkValidateReports(ctx context.Context, reportIDs []string, status string, rejectionReason *string) error {
	if status != "accept" && status != "reject" {
		return errors.New("status validasi tidak dikenali")
	}

	updateData := map[string]interface{}{
		"validation_status": status,
	}

	if rejectionReason != nil {
		updateData["rejection_reason"] = *rejectionReason
	} else {
		updateData["rejection_reason"] = nil
	}

	return r.db.WithContext(ctx).Model(&domain.InspectionReport{}).
		Where("id IN ?", reportIDs).
		Updates(updateData).Error
}

func (r *inspectionReportRepository) Update(ctx context.Context, reportID string, report *domain.InspectionReport) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("inspection_report_id = ?", reportID).Delete(&domain.ContainerDetail{}).Error; err != nil {
			return err
		}

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
			"validation_status": "pending",
		}

		if err := tx.Model(&domain.InspectionReport{}).Where("id = ?", reportID).Updates(updateData).Error; err != nil {
			return err
		}

		geomExpr := gorm.Expr("ST_SetSRID(ST_MakePoint(?::float, ?::float), 4326)", report.Longitude, report.Latitude)
		if err := tx.Model(&domain.InspectionReport{}).Where("id = ?", reportID).Update("geom", geomExpr).Error; err != nil {
			return err
		}

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
