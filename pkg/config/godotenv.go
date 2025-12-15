package config

import (
	"fmt"
	"os"
	"time"

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

	// Auth Settings
	PasetoSecret   string
	PasetoExp      time.Duration
	PasetoAudience string
	PasetoIssuer   string
}

// LoadConfig memuat konfigurasi dari file .env ke dalam struct
func LoadConfig() (*Config, error) {
	// 1. Muat file .env (jika ada)
	// Abaikan error jika file tidak ada, karena mungkin env vars
	// sudah diset di level OS/Container (berguna untuk Production nanti).
	_ = godotenv.Load()

	// Parse Duration secara eksplisit
	expStr := getEnv("PASETO_EXP_DURATION", "24h")
	expDuration, err := time.ParseDuration(expStr)
	if err != nil {
		return nil, fmt.Errorf("format durasi token salah: %w", err)
	}

	// 2. Baca ke dalam struct
	cfg := &Config{
		AppPort:        getEnv("APP_PORT", ":3000"),
		AppEnv:         getEnv("APP_ENV", "development"),
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBName:         os.Getenv("DB_NAME"),
		DBUser:         os.Getenv("DB_USERNAME"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBTimezone:     getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		DBSSLMode:      getEnv("DB_SSL_MODE", "disable"),
		PasetoSecret:   os.Getenv("PASETO_SECRET_KEY"),
		PasetoExp:      expDuration,
		PasetoAudience: getEnv("PASETO_AUDIENCE", "radar-jentik-app"),
		PasetoIssuer:   getEnv("PASETO_ISSUER", "radar-jentik-api"),
	}

	// 3. Validasi sederhana (Fail Fast)
	// Pastikan variabel kritikal tidak kosong
	if cfg.DBHost == "" || cfg.DBUser == "" {
		return nil, fmt.Errorf("konfigurasi database wajib (DB_HOST, DB_USERNAME) belum diisi")
	}
	// Validasi Panjang Key Paseto (Wajib 32 bytes untuk V2)
	if len(cfg.PasetoSecret) != 32 {
		return nil, fmt.Errorf("PASETO_SECRET_KEY wajib memiliki panjang tepat 32 karakter")
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
