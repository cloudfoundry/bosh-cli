package job

import (
	"fmt"
	"path/filepath"
	"testing"

	fsfakes "github.com/cloudfoundry/bosh-utils/system/fakes"
)

func TestJobTemplateSourcesStayWithinTemplates(t *testing.T) {
	for _, name := range []string{"../outside", "../../outside", "/outside"} {
		fs := fsfakes.NewFakeFileSystem()
		data := fmt.Sprintf("name: job\ntemplates:\n  %q: output\n", name)
		if err := fs.WriteFileString(filepath.Join("job", "spec"), data); err != nil {
			t.Fatal(err)
		}
		reader := DirReaderImpl{fs: fs}
		if _, _, err := reader.collectFiles("job"); err == nil {
			t.Fatalf("accepted job template path %q", name)
		}
	}
}
