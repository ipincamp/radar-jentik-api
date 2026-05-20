package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres/repositories"
	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
)

// Struct untuk mapping format GeoJSON
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	// Kita tidak lagi mem-parsing Geometry karena ERD baru hanya menyimpan Nama Desa
}

func main() {
	// 1. Parsing Arguments
	filePath := flag.String("file", "", "Path ke file GeoJSON (wajib)")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Harap sertakan path file: go run cmd/seeder/main.go -file=desa_kasegeran.json")
	}

	// 2. Load Config & Connect DB
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}

	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Gagal koneksi database: %v", err)
	}

	// 3. Init Village Repository yang baru
	villageRepo := repositories.NewVillageRepository(db)

	// 4. Baca File GeoJSON
	log.Printf("Membaca file: %s...", *filePath)
	byteValue, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file: %v", err)
	}

	// 5. Unmarshal JSON ke Struct
	var fc GeoJSONFeatureCollection
	if err := json.Unmarshal(byteValue, &fc); err != nil {
		log.Fatalf("Gagal parsing JSON: %v", err)
	}

	log.Printf("Ditemukan %d fitur desa. Memulai import...", len(fc.Features))

	// 6. Loop dan Simpan ke DB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	for _, f := range fc.Features {
		// Ambil Nama dari Properties (Sesuaikan "nm_kelurahan" dengan key asli di file GeoJSON Anda)
		nameIntf, ok := f.Properties["nm_kelurahan"]
		name := "Desa Tidak Diketahui"
		if ok {
			name = fmt.Sprintf("%v", nameIntf)
		}

		// Buat Domain Object Village
		newVillage := &domain.Village{
			Name: name,
		}

		// Simpan via Repository
		if err := villageRepo.Create(ctx, newVillage); err != nil {
			log.Printf("❌ Gagal menyimpan %s: %v", name, err)
		} else {
			log.Printf("✅ Berhasil menyimpan desa: %s (ID: %s)", newVillage.Name, newVillage.ID)
			successCount++
		}
	}

	log.Printf("🎉 Import selesai! %d/%d desa berhasil ditambahkan ke database.", successCount, len(fc.Features))
}
