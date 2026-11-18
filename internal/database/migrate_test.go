package database

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "ciranda.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	for _, table := range []string{"schema_migrations", "users"} {
		var count int
		err := db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM sqlite_schema
			WHERE type = 'table'
			  AND name = ?;
		`, table).Scan(&count)
		if err != nil {
			t.Fatalf("query %s table: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("%s table count = %d, want 1", table, count)
		}
	}

	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM schema_migrations
		WHERE version = 1;
	`).Scan(&count); err != nil {
		t.Fatalf("query migration version: %v", err)
	}
	if count != 1 {
		t.Fatalf("migration version 1 count = %d, want 1", count)
	}

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate database a second time: %v", err)
	}
}
