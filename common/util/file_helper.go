package util

import (
	"path/filepath"
	"strings"
)

func AbsolutifyPath(pathToManifest string, pathToFile string) string {
	if strings.HasPrefix(pathToFile, "http") {
		return pathToFile
	}

	if strings.HasPrefix(pathToFile, "file:///") || strings.HasPrefix(pathToFile, "/") {
		return pathToFile
	}

	if strings.HasPrefix(pathToFile, "file://~") || strings.HasPrefix(pathToFile, "~") {
		return pathToFile
	}

	var absPath string

	if strings.HasPrefix(pathToFile, "file://") {
		absPath += "file://"
	}

	absPath += filepath.Dir(pathToManifest)

	if !strings.HasSuffix(absPath, "/") {
		absPath += "/"
	}

	absPath += strings.Replace(pathToFile, "file://", "", 1)

	return absPath
}
