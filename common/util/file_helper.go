package util

import (
	"path/filepath"
	"strings"
)

func ParseFilePath(path string, filename string) string {
	if strings.Contains(filename, "file:///") || strings.Contains(filename, "file://~") {
		return filename
	}

	if strings.Contains(filename, "file://") {
		s := strings.Split(filepath.Dir(path), "/")
		s = append(s, strings.Replace(filename, "file://", "", 1))
		s = append([]string{"file:/"}, s...)
		return strings.Join(s, "/")
	}

	return filename
}
