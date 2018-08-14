package util

import (
	"path/filepath"
	"strings"
	"os"
)

func GetCurrentDirectory(file string) string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return file
	}
	return strings.Replace(dir, "\\", "/", -1) + "/" + file
}
