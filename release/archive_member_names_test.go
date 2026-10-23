package release_test

import (
	"strings"
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/release"
	"github.com/cloudfoundry/bosh-cli/v7/release/job"
	"github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fsfakes "github.com/cloudfoundry/bosh-utils/system/fakes"
)

func TestArchiveWriterRejectsInvalidMemberNames(t *testing.T) {
	fs := fsfakes.NewFakeFileSystem()
	writer := release.NewArchiveWriter(nil, fs, boshlog.NewLogger(boshlog.LevelNone))
	input := &releasefakes.FakeRelease{}
	input.JobsReturns([]*job.Job{job.NewJob(resource.NewResource("../outside", "", nil))})
	if _, err := writer.Write(input, nil); err == nil || !strings.Contains(err.Error(), "Invalid archive member name") {
		t.Fatalf("invalid archive member: %v", err)
	}
}
