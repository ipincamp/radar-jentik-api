package main

import (
	"context"
	"log"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
)

func main() {
	// 1. Load Konfigurasi & Koneksi Database
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}

	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Gagal koneksi database: %v", err)
	}

	// 2. Daftar 9 Wadah Default
	containers := []string{
		"Bak Kamar Mandi",
		"Tempayan",
		"Pecahan Botol/Air Kemasan",
		"Barang Bekas",
		"Kulkas/Dispenser",
		"Tandon Air",
		"Vas Bunga",
		"Pot Bunga",
		"Lain-lain",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	successCount := 0
	log.Println("Memulai proses injeksi data master jenis wadah...")

	// 3. Looping dan simpan ke database
	for _, name := range containers {
		container := domain.ContainerType{
			Name:     name,
			IsActive: true,
		}

		// Menggunakan FirstOrCreate dari GORM:
		// Jika wadah dengan nama tersebut sudah ada, lewati. Jika belum, buat baru.
		// Ini sangat aman walau script ini dijalankan berkali-kali.
		err := db.WithContext(ctx).Where(domain.ContainerType{Name: name}).FirstOrCreate(&container).Error
		if err != nil {
			log.Printf("❌ Gagal insert wadah '%s': %v", name, err)
		} else {
			log.Printf("✅ Wadah terdaftar: %s", name)
			successCount++
		}
	}

	log.Printf("🎉 Selesai! %d jenis wadah berhasil disiapkan di dalam database.", successCount)
}
