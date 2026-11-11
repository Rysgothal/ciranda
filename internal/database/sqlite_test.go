package database

import "testing"

func TestOpenSQLite(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open SQLite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Fatalf("query SQLite: %v", err)
	}
	if result != 1 {
		t.Fatalf("got %d, want 1", result)
	}
}
