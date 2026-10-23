package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

func TestDiagnosticFileModes(t *testing.T) {
	for _, existing := range []bool{false, true} {
		name := filepath.Join(t.TempDir(), "debug.log")
		if existing {
			if err := os.WriteFile(name, nil, 0644); err != nil {
				t.Fatal(err)
			}
		}
		logger := newSignalableFileLogger(name, boshlog.LevelDebug)
		logger.Debug("test", "private-diagnostic-value")
		info, err := os.Stat(name)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Fatalf("diagnostic mode %o", info.Mode().Perm())
		}
		data, err := os.ReadFile(name)
		if err != nil || !strings.Contains(string(data), "private-diagnostic-value") {
			t.Fatal("diagnostic not written")
		}
	}
}
