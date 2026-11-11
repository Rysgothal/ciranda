package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetLocalPath() string {
	ex, err := os.Executable()

	if err != nil {
		fmt.Print("INFO: local path not found")
		return ""
	}

	return filepath.Dir(ex)
}
