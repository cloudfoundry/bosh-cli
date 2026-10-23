package pkg

import (
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/release/pkg/manifest"
)

func TestPackagePatternsCheckedBeforeGlob(t *testing.T) {
	reader := DirReaderImpl{}
	for _, pattern := range []string{"../outside", "**/../../outside", "/outside", `..\outside`, "C:outside", ""} {
		if _, err := reader.applyFilesPattern(manifest.Manifest{Files: []string{pattern}}); err == nil {
			t.Fatalf("accepted package pattern %q", pattern)
		}
		if _, err := reader.applyExcludedFilesPattern(manifest.Manifest{ExcludedFiles: []string{pattern}}); err == nil {
			t.Fatalf("accepted exclusion pattern %q", pattern)
		}
	}
}
