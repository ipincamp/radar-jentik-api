package postgres

import (
	"fmt"
	"time"

	"github.com/ipincamp/radar-jentik-api/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnection membuat koneksi ke PostgreSQL menggunakan konfigurasi yang disediakan
func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	// 1. Data Source Name (DSN)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
		cfg.DBTimezone,
	)

	// 2. Konfigurasi GORM
	gormConfig := &gorm.Config{}

	// Aktifkan logger SQL jika di environment development/local agar query terlihat di terminal
	if cfg.AppEnv == "local" || cfg.AppEnv == "development" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Error) // Hanya log error di production
	}

	// 3. Buka Koneksi
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke database: %w", err)
	}

	// 4. Konfigurasi Connection Pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil instance sql.DB: %w", err)
	}

	// Set jumlah koneksi idle (menganggur) maksimal
	sqlDB.SetMaxIdleConns(10)

	// Set jumlah koneksi terbuka maksimal
	sqlDB.SetMaxOpenConns(100)

	// Set waktu maksimal koneksi boleh hidup
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
