package tarball_test

import (
	"path/filepath"
	"testing"

	tarball "github.com/cloudfoundry/bosh-cli/v7/installation/tarball"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fsfakes "github.com/cloudfoundry/bosh-utils/system/fakes"
)

func TestCacheDigestCannotChangeDirectory(t *testing.T) {
	base := t.TempDir()
	cache := tarball.NewCache(base, fsfakes.NewFakeFileSystem(), boshlog.NewLogger(boshlog.LevelNone))
	for _, digest := range []string{"../../outside", "x/../../outside", "x/../../../tmp/evil.tgz", "x/y", `..\..\outside`, "sha256:abc", "ordinary-digest"} {
		path := cache.Path(&fakeSource{url: "https://example.test/archive", sha1: digest})
		if filepath.Dir(path) != base {
			t.Fatalf("cache path left base directory: %s", path)
		}
	}
}
