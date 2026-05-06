package util

import (
	"fmt"
	gopath "path"
	"path/filepath"
	"strings"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

func AbsolutifyPath(pathToManifest string, pathToFile string, fs boshsys.FileSystem) (string, error) {
	if strings.HasPrefix(pathToFile, "http") {
		return pathToFile, nil
	}

	if strings.HasPrefix(pathToFile, "file:///") || strings.HasPrefix(pathToFile, "/") {
		return pathToFile, nil
	}

	if strings.HasPrefix(pathToFile, "file://~") {
		return pathToFile, nil
	}

	if strings.HasPrefix(pathToFile, "~") {
		return fs.ExpandPath(pathToFile)
	}

	var absPath string

	if !strings.HasPrefix(pathToFile, "file://") {
		absPath = filepath.Join(filepath.Dir(pathToManifest), pathToFile)
	} else {
		pathToFile = strings.Replace(pathToFile, "file://", "", 1)
		absPath = gopath.Join(gopath.Dir(pathToManifest), pathToFile)
		absPath = "file://" + absPath
	}

	return absPath, nil
}

// SafeJoinPath returns filepath.Join(base, untrusted), or an error if
// untrusted is not a local path.
func SafeJoinPath(base, untrusted string) (string, error) {
	if !filepath.IsLocal(untrusted) {
		return "", fmt.Errorf("path '%s' is not a safe local path", untrusted)
	}
	return filepath.Join(base, untrusted), nil
}
