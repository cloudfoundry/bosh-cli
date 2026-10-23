package templatescompiler_test

import (
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/release/job"
	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fsfakes "github.com/cloudfoundry/bosh-utils/system/fakes"
)

func TestRenderedJobNameStaysWithinArchive(t *testing.T) {
	fs := fsfakes.NewFakeFileSystem()
	logger := boshlog.NewLogger(boshlog.LevelNone)
	list := templatescompiler.NewRenderedJobList()
	input := job.NewJob(resource.NewResource("../outside", "", nil))
	list.Add(templatescompiler.NewRenderedJob(*input, "input", fs, logger))
	compressor := templatescompiler.NewRenderedJobListCompressor(fs, nil, nil, logger)
	if _, err := compressor.Compress(list); err == nil {
		t.Fatal("accepted rendered job outside archive directory")
	}
}
