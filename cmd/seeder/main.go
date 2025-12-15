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
	// Gunakan json.RawMessage agar Geometry tetap berupa string JSON mentah
	// karena ST_GeomFromGeoJSON di PostGIS membutuhkan format JSON utuh.
	Geometry json.RawMessage `json:"geometry"`
}

func main() {
	// 1. Parsing Arguments
	filePath := flag.String("file", "", "Path ke file GeoJSON (wajib)")
	areaType := flag.String("type", "desa", "Tipe area (contoh: desa, rw, kecamatan)")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Harap sertakan path file: go run cmd/seeder/main.go -file=desa_panusupan.txt")
	}

	// 2. Load Config & Database
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}

	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Gagal koneksi database: %v", err)
	}

	// 3. Init Repository
	areaRepo := repositories.NewAreaRepo(db)

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

	log.Printf("Ditemukan %d fitur area. Memulai import...", len(fc.Features))

	// 6. Loop dan Simpan ke DB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	for _, f := range fc.Features {
		// Ambil Nama dari Properties (sesuaikan dengan key di file Anda: "nm_kelurahan")
		nameIntf, ok := f.Properties["nm_kelurahan"]
		name := "Unknown Area"
		if ok {
			name = fmt.Sprintf("%v", nameIntf)
		}

		// Konversi Geometry JSON ke String
		geomString := string(f.Geometry)

		// Buat Domain Object
		newArea := &domain.Area{
			Name: name,
			Type: *areaType,
		}

		// Simpan via Repository
		// Repository akan membungkus geomString dengan ST_GeomFromGeoJSON
		if err := areaRepo.Save(ctx, newArea, geomString); err != nil {
			log.Printf("[ERROR] Gagal menyimpan area '%s': %v", name, err)
		} else {
			log.Printf("[SUCCESS] Berhasil import: %s", name)
			successCount++
		}
	}

	log.Printf("Selesai! %d/%d data berhasil diimport.", successCount, len(fc.Features))
}
