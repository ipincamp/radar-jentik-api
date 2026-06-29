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
	Geometry   json.RawMessage        `json:"geometry"`
}

func main() {
	// 1. Parsing Arguments
	filePath := flag.String("file", "", "Path ke file GeoJSON (wajib)")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Harap sertakan path file: go run cmd/seeder/main.go -file=33.02_kelurahan.geojson")
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

	// 3. Baca File GeoJSON
	log.Printf("Membaca file: %s...", *filePath)
	byteValue, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file: %v", err)
	}

	// 4. Unmarshal JSON ke Struct
	var fc GeoJSONFeatureCollection
	if err := json.Unmarshal(byteValue, &fc); err != nil {
		log.Fatalf("Gagal parsing JSON: %v", err)
	}

	log.Printf("Ditemukan total %d fitur spasial di file. Memulai penyaringan desa...", len(fc.Features))

	// 5. Daftar Desa Target di Kecamatan Cilongok
	// Menggunakan Map (Hash Table) untuk pencarian cepat (O(1))
	targetVillages := map[string]bool{
		"Batuanten":    true,
		"Cipete":       true,
		"Jatisaba":     true,
		"Kasegeran":    true,
		"Langgongsari": true,
		"Pageraji":     true,
		"Panusupan":    true,
		"Pejogol":      true,
		"Sudimara":     true,
	}

	successCount := 0

	// Gunakan timeout agar aman
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 6. Loop dan Simpan ke DB
	for _, f := range fc.Features {
		// Ambil Nama Kelurahan/Desa dari file GeoJSON
		nameIntf, ok := f.Properties["nm_kelurahan"]
		if !ok {
			continue
		}
		name := fmt.Sprintf("%v", nameIntf)

		// FILTERING: Lewati (skip) jika nama desa tidak ada di dalam daftar targetVillages
		if !targetVillages[name] {
			continue
		}

		// Konversi data geometri mentah (RawMessage) ke dalam format String JSON
		geoJSONStr := string(f.Geometry)

		// 7. INSERT via Raw Query (Bypass Repository khusus untuk Seeder)
		// Menggunakan ST_GeomFromGeoJSON untuk menerjemahkan string JSON menjadi titik Geometri spasial
		// Menggunakan ST_Multi untuk memastikan tipe datanya MultiPolygon
		// Menggunakan ST_Force2D untuk membuang dimensi Z (ketinggian) dari file GeoJSON
		query := `
			INSERT INTO villages (id, name, boundary, created_at, updated_at)
			VALUES (uuid_generate_v4(), ?, ST_Force2D(ST_Multi(ST_SetSRID(ST_GeomFromGeoJSON(?), 4326))), NOW(), NOW())
		`

		if err := db.Exec(query, name, geoJSONStr).Error; err != nil {
			log.Printf("❌ Gagal menyimpan desa %s: %v", name, err)
		} else {
			log.Printf("✅ Berhasil menyimpan desa target: %s", name)
			successCount++
		}
	}

	log.Printf("🎉 Import selesai! %d desa target berhasil ditambahkan ke database.", successCount)
}
