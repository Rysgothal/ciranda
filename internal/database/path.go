package database

import (
	"fmt"
	"os"
	"path/filepath"
)

func DefaultPath() (string, error) {
	dataDir := "data"

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", fmt.Errorf("cannot create data directory: %w", err)
	}

	return filepath.Join(dataDir, "ciranda.db"), nil
}
