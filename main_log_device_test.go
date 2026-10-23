package main

import (
	"os"
	"testing"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
)

func TestRestrictLogFileLeavesDevicesUnchanged(t *testing.T) {
	fs := system.NewOsFileSystem(boshlog.NewLogger(boshlog.LevelNone))
	file, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close() //nolint:errcheck
	before, err := os.Stat(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	if err := restrictLogFile(file, fs); err != nil {
		t.Fatalf("device log path rejected: %v", err)
	}
	after, err := os.Stat(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode() != after.Mode() {
		t.Fatalf("device mode changed from %v to %v", before.Mode(), after.Mode())
	}
}
