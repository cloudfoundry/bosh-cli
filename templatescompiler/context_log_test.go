package templatescompiler_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/release/job"
	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/property"
	uuidfakes "github.com/cloudfoundry/bosh-utils/uuid/fakes"
)

func TestContextDiagnosticExcludesProperties(t *testing.T) {
	var output bytes.Buffer
	logger := boshlog.New(boshlog.LevelDebug, log.New(&output, "", 0))
	releaseJob := job.NewJob(resource.NewResource("job", "", nil))
	context := templatescompiler.NewJobEvaluationContext(*releaseJob, nil, nil, property.Map{"password": "private-context-value"}, "deployment", "", uuidfakes.NewFakeGenerator(), logger)
	data, err := context.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "private-context-value") {
		t.Fatal("context value lost")
	}
	if strings.Contains(output.String(), "private-context-value") {
		t.Fatal("properties appeared in context log")
	}
}
