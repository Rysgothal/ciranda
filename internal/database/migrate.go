package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if err = createSchemaMigrationsTable(ctx, db, migrations); err != nil {
		return err
	}

	return applySchemaMigrations(ctx, db, migrations)
}

func loadMigrations() ([]string, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	migrations := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}

		migrations = append(migrations, string(content))
	}

	return migrations, nil
}

func createSchemaMigrationsTable(ctx context.Context, db *sql.DB, migrations []string) error {
	const querySchemaMigrations = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		);
	`

	if _, err := db.ExecContext(ctx, querySchemaMigrations); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	return nil
}

func applySchemaMigrations(ctx context.Context, db *sql.DB, migrations []string) error {
	for i, migration := range migrations {
		version := i + 1
		if err := execMigration(ctx, db, version, migration); err != nil {
			return fmt.Errorf("apply migration %d: %w", version, err)
		}
	}

	if err := applyForeignKeys(ctx, db); err != nil {
		return err
	}

	return nil
}

func applyForeignKeys(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`)
	if err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	return nil
}

func execMigration(ctx context.Context, db *sql.DB, version int, migration string) error {
	const queryVerifyMigrationApplied = "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)"
	const queryApplyMigration = "INSERT INTO schema_migrations (version, applied_at) VALUES (?, datetime('now'))"

	var applied bool
	if err := db.QueryRowContext(ctx, queryVerifyMigrationApplied, version).Scan(&applied); err != nil {
		return fmt.Errorf("verify migration %d: %w", version, err)
	}
	if applied {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", version, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, migration); err != nil {
		return fmt.Errorf("exec migration %d: %w", version, err)
	}
	if _, err := tx.ExecContext(ctx, queryApplyMigration, version); err != nil {
		return fmt.Errorf("record migration %d: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", version, err)
	}

	return nil
}
