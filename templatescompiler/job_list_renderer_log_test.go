package templatescompiler_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/property"

	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/templatescompilerfakes"
)

func TestJobListRendererLogsExcludeProperties(t *testing.T) {
	var output bytes.Buffer
	logger := boshlog.New(boshlog.LevelDebug, log.New(&output, "", 0))
	renderer := templatescompiler.NewJobListRenderer(&templatescompilerfakes.FakeJobRenderer{}, logger)
	jobProperties := property.Map{"password": "private-job-value"}
	globalProperties := property.Map{"password": "private-global-value"}
	if _, err := renderer.Render(nil, nil, jobProperties, globalProperties, "deployment", ""); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private-job-value") || strings.Contains(output.String(), "private-global-value") {
		t.Fatal("properties appeared in job list log")
	}
	if !strings.Contains(output.String(), "deployment") {
		t.Fatal("deployment name missing from job list log")
	}
}
