package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres/repositories"
	internalHttp "github.com/ipincamp/radar-jentik-api/internal/adapters/driving/http"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driving/http/handlers"
	"github.com/ipincamp/radar-jentik-api/internal/core/services"
	"github.com/ipincamp/radar-jentik-api/pkg/auth"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
)

func main() {
	// 1. Load Konfigurasi
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal memuat konfigurasi: %v", err)
	}
	log.Printf("Aplikasi berjalan di environment: %s", cfg.AppEnv)

	// 2. Inisialisasi Database (Driven Adapter)
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database tidak merespon: %v", err)
	}
	log.Println("Berhasil terhubung ke database PostgreSQL")

	// 3. Dependency Injection (DI) Container
	tokenManager := auth.NewTokenManager(cfg)

	// A. Init Repositories
	userRepo := repositories.NewUserRepo(db)
	inspectionReportRepo := repositories.NewInspectionReportRepository(db)
	villageRepo := repositories.NewVillageRepository(db) // Daftarkan repo desa

	// B. Init Services
	authService := services.NewAuthService(userRepo, tokenManager)
	inspectionReportService := services.NewInspectionReportService(inspectionReportRepo)
	villageService := services.NewVillageService(villageRepo) // Daftarkan service desa
	idwService := services.NewIDWService()

	// C. Init Handlers
	authHandler := handlers.NewAuthHandler(authService)
	inspectionReportHandler := handlers.NewInspectionReportHandler(inspectionReportService)
	villageHandler := handlers.NewVillageHandler(villageService) // Daftarkan handler desa
	idwHandler := handlers.NewIDWHandler(idwService)

	// 4. Inisialisasi HTTP Adapter (Fiber Server) dengan semua handler lengkap
	server := internalHttp.NewServer(cfg, authHandler, inspectionReportHandler, villageHandler, idwHandler)

	// Channel untuk menangkap signal interrupt (seperti Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Jalankan server di goroutine terpisah agar tidak memblokir quit channel
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)
		log.Printf("Server is starting on %s...", addr)
		if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Fatal error starting server: %v", err)
		}
	}()

	// Menunggu signal interrupt
	<-quit
	log.Printf("Shutting down server gracefully...")

	// Konteks untuk shutdown dengan timeout (10 detik)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Fatal error during server shutdown: %v", err)
	}

	log.Printf("Server stopped gracefully")
}
