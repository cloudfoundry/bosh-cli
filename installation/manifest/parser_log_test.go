package manifest_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	"github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	"github.com/cloudfoundry/bosh-cli/v7/installation/manifest/manifestfakes"
	releasemanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
	uuidfakes "github.com/cloudfoundry/bosh-utils/uuid/fakes"
)

func TestInstallationParserLogsExcludeContents(t *testing.T) {
	var output bytes.Buffer
	logger := boshlog.New(boshlog.LevelDebug, log.New(&output, "", 0))
	file := filepath.Join(t.TempDir(), "manifest.yml")
	contents := "name: deployment\ncloud_provider:\n  templates:\n  - name: cpi\n    release: cpi-release\n  mbus: https://user:private-password@host\n"
	if err := os.WriteFile(file, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	parser := manifest.NewParser(system.NewOsFileSystem(logger), uuidfakes.NewFakeGenerator(), logger, &manifestfakes.FakeValidator{})
	if _, err := parser.Parse(file, boshtpl.StaticVariables{}, nil, releasemanifest.Manifest{}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private-password") {
		t.Fatal("installation credentials appeared in logs")
	}
}
