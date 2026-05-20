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
	inspectionReportRepo := repositories.NewInspectionReportRepository(db)

	// C. Init Service (Inject Repo & TokenManager)
	authService := services.NewAuthService(userRepo, tokenManager)
	inspectionReportService := services.NewInspectionReportService(inspectionReportRepo)

	// D. Init Handler
	authHandler := handlers.NewAuthHandler(authService)
	inspectionReportHandler := handlers.NewInspectionReportHandler(inspectionReportService)

	// 4. Inisialisasi HTTP Adapter (Fiber)
	server := internalHttp.NewServer(cfg, authHandler, inspectionReportHandler)

	// Channel untuk menangkap signal interrupt (seperti Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Jalankan server di goroutine terpisah agar tidak memblokir quit channel
	go func() {
		addr := fmt.Sprintf("0.0.0.0:%s", cfg.AppPort)
		log.Printf("Server is starting on %s...", addr)
		if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Fatal error starting server: %v", err)
		}
	}()

	// Menunggu signal interrupt
	<-quit
	log.Printf("Shutting down server gracefully...")

	// Konteks untuk shutdown dengan timeout (misal 10 detik)
	// Agar server tidak langsung mati dan bisa menyelesaikan request yang sedang berjalan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Printf("Server exited properly")
}
