package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Application settings
	AppPort string
	AppEnv  string

	// Database settings
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBTimezone string
	DBSSLMode  string
}

// LoadConfig memuat konfigurasi dari file .env ke dalam struct
func LoadConfig() (*Config, error) {
	// 1. Muat file .env (jika ada)
	// Abaikan error jika file tidak ada, karena mungkin env vars
	// sudah diset di level OS/Container (berguna untuk Production nanti).
	_ = godotenv.Load()

	// 2. Baca ke dalam struct
	cfg := &Config{
		AppPort:    getEnv("APP_PORT", ":3000"),
		AppEnv:     getEnv("APP_ENV", "development"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USERNAME"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBTimezone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	// 3. Validasi sederhana (Fail Fast)
	// Pastikan variabel kritikal tidak kosong
	if cfg.DBHost == "" || cfg.DBUser == "" {
		return nil, fmt.Errorf("konfigurasi database wajib (DB_HOST, DB_USERNAME) belum diisi")
	}

	return cfg, nil
}

// Helper untuk mengambil env dengan nilai default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
