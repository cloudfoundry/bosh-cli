package releasedir

import (
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

func TestReleaseFileComponentsBeforeFilesystem(t *testing.T) {
	index := FSReleaseIndex{}
	for _, name := range []string{"../outside", "..", `..\outside`, "/outside"} {
		if _, err := index.LastVersion(name); err == nil {
			t.Fatalf("accepted release index name %q", name)
		}
		if err := index.Add(manifest.Manifest{Name: name, Version: "1"}); err == nil {
			t.Fatalf("accepted release name %q", name)
		}
		if err := index.Add(manifest.Manifest{Name: "release", Version: name}); err == nil {
			t.Fatalf("accepted release version %q", name)
		}
		pkg := boshpkg.NewPackage(resource.NewResource(name, "", nil), nil)
		if err := (FSReleaseDir{}).writeVendoredPackage(pkg); err == nil {
			t.Fatalf("accepted vendored package %q", name)
		}
	}
}
