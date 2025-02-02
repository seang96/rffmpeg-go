package processor

import (
	"embed"
	"fmt"

	"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/aleksasiriski/rffmpeg-go/migrate"
)

type datastore struct {
	*sql.DB
	dbType string
}

//go:embed migrations/sqlite
var migrationsSqlite embed.FS

//go:embed migrations/postgres
var migrationsPostgres embed.FS

func newDatastore(db *sql.DB, dbType string, mg *migrate.Migrator) (*datastore, error) {
	switch dbType {
	case "sqlite":
		{
			// migrations/sqlite
			if err := mg.Migrate(&migrationsSqlite, "processor"); err != nil {
				return nil, fmt.Errorf("migrate: %w", err)
			}
		}
	case "postgres":
		{
			// migrations/postgres
			if err := mg.Migrate(&migrationsPostgres, "processor"); err != nil {
				return nil, fmt.Errorf("migrate: %w", err)
			}
		}
	default:
		return &datastore{}, fmt.Errorf("incorrect database type")
	}
	return &datastore{db, dbType}, nil
}

func sqlSelectVersion(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return "sqlite", nil
	case "postgres":
		return `SELECT version()`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectVersion() (string, error) {
	sqlSelectVersion, err := sqlSelectVersion(store.dbType)
	if err != nil {
		return "", err
	}

	version := "sqlite"
	if store.dbType != "sqlite" {
		row := store.QueryRow(sqlSelectVersion)
		err = row.Scan(&version)
		if err != nil {
			return version, fmt.Errorf("select version: %w", err)
		}
	}

	return version, nil
}
