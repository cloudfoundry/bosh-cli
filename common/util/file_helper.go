package util

import (
	"path/filepath"
	"strings"
)

func ParseFilePath(path string, filename string) string {
	if strings.HasPrefix(filename, "file:///") || strings.HasPrefix(filename, "file://~") ||
		strings.HasPrefix(filename, "http") || strings.HasPrefix(filename, "/") {
		return filename
	}

	s := strings.Split(filepath.Dir(path), "/")
	if strings.HasPrefix(filename, "file://") {
		s = append(s, strings.Replace(filename, "file://", "", 1))
		s = append([]string{"file:/"}, s...)
	} else {
		s = append(s, filename)
	}

	return strings.Join(s, "/")
}
