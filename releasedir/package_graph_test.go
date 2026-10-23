package releasedir

import (
	"testing"

	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
)

func TestPackageCollectionHandlesCyclesAndSharedDependencies(t *testing.T) {
	a, b, c := &boshpkg.Package{}, &boshpkg.Package{}, &boshpkg.Package{}
	a.Dependencies = []*boshpkg.Package{b, c}
	b.Dependencies = []*boshpkg.Package{a, c}
	c.Dependencies = []*boshpkg.Package{c}
	all := map[*boshpkg.Package]struct{}{}
	FSReleaseDir{}.collectDependentPackages(a, all)
	if len(all) != 3 {
		t.Fatalf("collected %d packages", len(all))
	}
	for _, pkg := range []*boshpkg.Package{a, b, c} {
		if _, found := all[pkg]; !found {
			t.Fatal("reachable package missing")
		}
	}
}
