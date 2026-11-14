package database

import (
	"context"
	"database/sql"
	"fmt"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	return createSchemaMigrationsTable(ctx, db)
}

func createSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		);
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	return nil
}
