package database

import (
	"context"
	"testing"
)

func TestMigrateCreatesSchemaMigrationsTable(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	var count int

	err = db.QueryRowContext(context.Background(), `
        SELECT COUNT(*)
        FROM sqlite_schema
        WHERE type = 'table'
          AND name = 'schema_migrations';
	`).Scan(&count)

	if err != nil {
		t.Fatalf("query schema_migrations table: %v", err)
	}

	if count != 1 {
		t.Fatalf("schema_migrations table count = %d, want 1", count)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})
}
