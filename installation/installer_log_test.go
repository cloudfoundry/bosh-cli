package installation_test

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"

	"github.com/cloudfoundry/bosh-cli/v7/installation"
	"github.com/cloudfoundry/bosh-cli/v7/installation/blobextract/blobextractfakes"
	"github.com/cloudfoundry/bosh-cli/v7/installation/installationfakes"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	fakebiui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

func TestInstallerLogsExcludeManifestContents(t *testing.T) {
	var output bytes.Buffer
	logger := boshlog.New(boshlog.LevelDebug, log.New(&output, "", 0))
	resolver := &installationfakes.FakeJobResolver{}
	resolver.FromReturns(nil, errors.New("stop after logging"))
	installer := installation.NewInstaller(
		installation.NewTarget("installation-path", ""),
		&installationfakes.FakeJobRenderer{},
		resolver,
		&installationfakes.FakePackageCompiler{},
		&blobextractfakes.FakeExtractor{},
		logger,
	)
	manifest := biinstallmanifest.Manifest{
		Name:       "installation",
		Properties: biproperty.Map{"password": "private-property-value"},
		Mbus:       "https://mbus:private-mbus-value@host:6868",
	}
	if _, err := installer.Install(manifest, fakebiui.NewFakeStage()); err == nil {
		t.Fatal("expected resolver error")
	}
	if strings.Contains(output.String(), "private-property-value") || strings.Contains(output.String(), "private-mbus-value") {
		t.Fatal("installation manifest contents appeared in logs")
	}
}
