package main

import (
	"log"

	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres/repositories"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driving/http"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driving/http/handlers"
	"github.com/ipincamp/radar-jentik-api/internal/core/services"
	"github.com/ipincamp/radar-jentik-api/pkg/auth"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
)

func main() {
	// 1. Load Konfigurasi
	cfg, err := config.LoadConfig()
	if err != nil {
		// Log fatal akan menghentikan aplikasi jika config gagal
		log.Fatalf("Gagal memuat konfigurasi: %v", err)
	}
	// Tampilkan info (hanya untuk debug)
	log.Printf("Aplikasi berjalan di environment: %s", cfg.AppEnv)
	log.Printf("Database target: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	// 2. Inisialisasi Database (Driven Adapter)
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	// Cek koneksi database
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database tidak merespon: %v", err)
	} else {
		log.Println("Berhasil terhubung ke database PostgreSQL")
	}

	// 3. Dependency Injection (DI) Container
	// A. Init Token Manager (Auth Utility)
	tokenManager := auth.NewTokenManager(cfg)

	// B. Init Repository
	userRepo := repositories.NewUserRepo(db)
	reportRepo := repositories.NewReportRepo(db)
	areaRepo := repositories.NewAreaRepo(db)

	// C. Init Service (Inject Repo & TokenManager)
	authService := services.NewAuthService(userRepo, tokenManager)
	reportService := services.NewReportService(reportRepo)
	areaService := services.NewAreaService(areaRepo)

	// D. Init Handler
	authHandler := handlers.NewAuthHandler(authService)
	reportHandler := handlers.NewReportHandler(reportService)
	areaHandler := handlers.NewAreaHandler(areaService)

	// 4. Inisialisasi HTTP Adapter (Fiber)
	httpServer := http.NewServer(cfg, authHandler, reportHandler, areaHandler)

	// 5. Jalankan Server
	log.Printf("Menjalankan server di port %s pada environment %s", cfg.AppPort, cfg.AppEnv)
	if err := httpServer.Run(); err != nil {
		log.Fatalf("Server gagal berjalan: %v", err)
	}

}
