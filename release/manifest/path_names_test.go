package manifest_test

import (
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	fsfakes "github.com/cloudfoundry/bosh-utils/system/fakes"
)

func TestManifestFileComponents(t *testing.T) {
	for _, data := range []string{
		"name: ../outside\nversion: '1'\n",
		"name: release\nversion: ../outside\n",
		"jobs:\n- name: ../outside\n",
		"packages:\n- name: ../outside\n",
		"compiled_packages:\n- name: ../outside\n",
	} {
		fs := fsfakes.NewFakeFileSystem()
		if err := fs.WriteFileString("manifest", data); err != nil {
			t.Fatal(err)
		}
		if _, err := manifest.NewManifestFromPath("manifest", fs); err == nil {
			t.Fatal("accepted invalid release file component")
		}
	}
}
