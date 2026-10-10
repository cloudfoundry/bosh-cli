//go:build !windows

package system

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func (fs *osFileSystem) homeDir(username string) (string, error) {
	homeDir, err := fs.runCommand(fmt.Sprintf("echo ~%s", username))
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Shelling out to get user '%s' home directory", username)
	}
	if strings.HasPrefix(homeDir, "~") {
		return "", bosherr.Errorf("Failed to get user '%s' home directory", username)
	}
	return homeDir, nil
}

func (fs *osFileSystem) chown(path, owner string) error {
	if owner == "" {
		return errors.New("failed to lookup user ''")
	}

	var group string
	var err error

	ownerSplit := strings.Split(owner, ":")
	user := ownerSplit[0]

	if len(ownerSplit) <= 1 {
		group, err = runWithoutShell("id", "-g", "--", user)
		if err != nil {
			return bosherr.WrapErrorf(err, "failed to lookup user '%s'", user)
		}
	} else {
		group = ownerSplit[1]
	}

	_, err = runWithoutShell("chown", "--", user+":"+group, path)
	if err != nil {
		return bosherr.WrapError(err, "failed to chown")
	}

	return nil
}

// runWithoutShell passes each argument as is, so a path or user name cannot run shell commands.
func runWithoutShell(name string, args ...string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (fs *osFileSystem) symlinkPaths(oldPath, newPath string) (old, new string, err error) {
	return oldPath, newPath, nil
}
