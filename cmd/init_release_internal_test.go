package cmd

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
)

func TestInitReleaseDir(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working directory: %s", err)
	}

	tests := []struct {
		name     string
		dir      DirOrCWDArg
		release  string
		expected string
	}{
		{
			name:     "no name and no dir uses the working directory",
			dir:      DirOrCWDArg{},
			expected: cwd,
		},
		{
			name:     "name without dir creates a subdirectory",
			dir:      DirOrCWDArg{},
			release:  "my-release",
			expected: filepath.Join(cwd, "my-release"),
		},
		{
			name:     "explicit dir is used as the release root",
			dir:      DirOrCWDArg{Path: "/releases/here"},
			release:  "my-release",
			expected: "/releases/here",
		},
		{
			name:     "explicit current dir is not turned into a subdirectory",
			dir:      DirOrCWDArg{Path: cwd},
			release:  "my-release",
			expected: cwd,
		},
		{
			name:     "explicit dir without a name is used as-is",
			dir:      DirOrCWDArg{Path: "/releases/here"},
			expected: "/releases/here",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := initReleaseDir(test.dir, test.release)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if dir.Path != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, dir.Path)
			}
		})
	}
}
