package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	var err error
	var db *sql.DB

	db, err = sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("cannot Open database: %w", err)
	}

	return db, nil
}
