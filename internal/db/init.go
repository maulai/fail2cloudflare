package db

import (
	"database/sql"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func Init(dbPath string) (*sql.DB, error) {
	// Open database
	sqlitedb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	sqlitedb.Exec("PRAGMA busy_timeout = 5000;")
	sqlitedb.Exec("PRAGMA journal_mode = 'WAL';")

	return sqlitedb, nil
}

func Migrate(sqlitedb *sql.DB) error {
	// Apply migrations
	goose.SetBaseFS(MigrationFS)
	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}
	if _, err := goose.EnsureDBVersion(sqlitedb); err != nil {
		return err
	}
	if err := goose.Up(sqlitedb, "migrations"); err != nil {
		return err
	}

	return nil
}
