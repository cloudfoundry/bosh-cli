package manifest_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/template"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
)

func TestDeploymentParserLogsExcludeContents(t *testing.T) {
	var output bytes.Buffer
	logger := boshlog.New(boshlog.LevelDebug, log.New(&output, "", 0))
	parser := manifest.NewParser(system.NewOsFileSystem(logger), logger)
	input := template.NewInterpolatedTemplate([]byte("name: deployment\nproperties:\n  password: private-password\n"), "")
	if _, err := parser.Parse(input, "manifest.yml"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private-password") {
		t.Fatal("deployment credentials appeared in logs")
	}
}
