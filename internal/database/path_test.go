package database

import "testing"

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()

	if err != nil {
		t.Fatalf("get default path: %v", err)
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open SQLite: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close SQLite: %v", err)
	}
}
