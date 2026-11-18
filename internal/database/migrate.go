package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/001_create_tables.sql
var migration001 string

func Migrate(ctx context.Context, db *sql.DB) error {
	return createSchemaMigrationsTable(ctx, db)
}

func createSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	const querySchemaMigrations = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		);
	`
	_, err := db.ExecContext(ctx, querySchemaMigrations)
	if err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	if err = applySchemaMigration(ctx, db, 1); err != nil {
		return fmt.Errorf("apply migration %d: %w", 1, err)
	}

	return nil
}

func applySchemaMigration(ctx context.Context, db *sql.DB, version int) error {
	const queryVerifyMigrationApplied = "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)"
	const queryApplyMigration = "INSERT INTO schema_migrations (version, applied_at) VALUES (?, datetime('now'))"

	var err error
	var applied bool

	err = db.QueryRowContext(ctx, queryVerifyMigrationApplied, version).Scan(&applied)
	if err != nil {
		return fmt.Errorf("verify migration %d: %w", version, err)
	}

	if applied {
		return nil
	}

	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, getMigration(version))
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec migration %d: %w", version, err)
	}

	_, err = tx.ExecContext(ctx, queryApplyMigration, version)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply migration %d: %w", version, err)
	}

	return tx.Commit()
}

func getMigration(version int) string {
	switch version {
	case 1:
		return migration001
	default:
		return ""
	}
}
