package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
)

// GetMigrations mengembalikan daftar semua migrasi yang terdaftar
// Urutan dalam slice ini PENTING. Jangan ubah urutan migrasi lama.
func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		// Migrasi baru...
	}
}
