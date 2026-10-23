package erbrenderer

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
)

type privateContext struct{}

func (privateContext) MarshalJSON() ([]byte, error) {
	return []byte(`{"password":"private-context-value"}`), nil
}
func TestContextFileContentsExcludedFromLogs(t *testing.T) {
	var output bytes.Buffer
	logger := boshlog.New(boshlog.LevelDebug, log.New(&output, "", 0))
	renderer := erbRenderer{fs: system.NewOsFileSystem(logger)}
	file := filepath.Join(t.TempDir(), "context")
	if err := renderer.writeContext(file, privateContext{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private-context-value") {
		t.Fatal("context content appeared in logs")
	}
	contents, err := os.ReadFile(file)
	if err != nil || !strings.Contains(string(contents), "private-context-value") {
		t.Fatal("context file not preserved")
	}
}
