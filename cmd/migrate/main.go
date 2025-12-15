package main

import (
	"flag"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres"
	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres/migrations"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
	"gorm.io/gorm"
)

func main() {
	// 1. Parse Command Line Flags
	action := flag.String("action", "up", "Pilihan: up (migrate), down (rollback last), fresh (reset db)")
	flag.Parse()

	// 2. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}

	// 3. Connect DB
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Gagal koneksi database: %v", err)
	}

	// 4. Inisialisasi Gormigrate dengan Registry
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations.GetMigrations())

	log.Printf("Menjalankan aksi migrasi: %s", *action)

	// 5. Eksekusi Aksi
	switch *action {
	case "up":
		if err = m.Migrate(); err != nil {
			log.Fatalf("Gagal Migrate: %v", err)
		}
		log.Println("Migrasi Berhasil!")

	case "down":
		if err = m.RollbackLast(); err != nil {
			log.Fatalf("Gagal Rollback: %v", err)
		}
		log.Println("Rollback Berhasil!")

	case "fresh":
		// Peringatan: Hati-hati menggunakan ini!
		if err := freshDatabase(db, m); err != nil {
			log.Fatalf("Gagal Fresh Database: %v", err)
		}
		log.Println("Database Fresh Berhasil (Reset Total)!")

	default:
		log.Fatal("Action tidak valid. Gunakan: -action=[up|down|fresh]")
	}
}

// freshDatabase menghapus schema public dan menjalankan migrasi ulang
func freshDatabase(db *gorm.DB, m *gormigrate.Gormigrate) error {
	// 1. Drop Schema Public (Menghapus semua tabel & fungsi)
	// Hanya support PostgreSQL
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DROP SCHEMA public CASCADE").Error; err != nil {
			return err
		}
		if err := tx.Exec("CREATE SCHEMA public").Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 2. Jalankan Migrate dari awal
	return m.Migrate()
}
