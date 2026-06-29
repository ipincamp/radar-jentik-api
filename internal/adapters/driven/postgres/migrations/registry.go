package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
)

func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		CreateVillagesTable(),
		CreateUsersTable(),
		CreateContainerTypesTable(),
		CreateInspectionReportsTable(),
		CreateContainerDetailsTable(),
	}
}
