package main

import (
	"log"

	"github.com/ipincamp/radar-jentik-api/internal/adapters/driving/http"
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

	// 2. Inisialisasi HTTP Adapter (Fiber)
	httpServer := http.NewServer(cfg)

	// 3. Jalankan Server
	log.Printf("Menjalankan server di port %s pada environment %s", cfg.AppPort, cfg.AppEnv)
	if err := httpServer.Run(); err != nil {
		log.Fatalf("Server gagal berjalan: %v", err)
	}

}
