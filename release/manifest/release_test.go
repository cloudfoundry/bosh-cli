package manifest_test

import (
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	. "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
)

var _ = Describe("NewManifestFromPath", func() {
	var (
		fs *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
	})

	It("decodes release with base64 job/package sha1s and fingerprints", func() {
		err := fs.WriteFileString("/release.yml", `---
name: release
version: version
commit_hash: commit
uncommitted_changes: true

jobs:
- name: job1
  sha1: !binary |-
    am9iMS1zaGEx
  fingerprint: !binary |-
    am9iMS1mcA==
  version: !binary |-
    am9iMS12ZXJzaW9u

packages:
- name: pkg1
  sha1: !binary |-
    cGtnMS1zaGEx
  fingerprint: !binary |-
    cGtnMS1mcA==
  version: !binary |-
    cGtnMS12ZXJzaW9u

compiled_packages:
- name: compiled-pkg
  fingerprint: compiled-fingerprint
  sha1: compiled-sha1
`)
		Expect(err).ToNot(HaveOccurred())

		manifest, err := NewManifestFromPath("/release.yml", fs)
		Expect(err).NotTo(HaveOccurred())

		Expect(manifest.Jobs[0].SHA1).To(Equal("job1-sha1"))
		Expect(manifest.Jobs[0].Fingerprint).To(Equal("job1-fp"))
		Expect(manifest.Jobs[0].Version).To(Equal("job1-version"))

		Expect(manifest.Packages[0].SHA1).To(Equal("pkg1-sha1"))
		Expect(manifest.Packages[0].Fingerprint).To(Equal("pkg1-fp"))
		Expect(manifest.Packages[0].Version).To(Equal("pkg1-version"))
		Expect(manifest.CompileTimePackages).To(BeEmpty())
		Expect(manifest.CompiledPkgs[0].OriginalFingerprint).To(BeEmpty())
	})

	It("decodes compile-time package metadata and original compiled package fingerprints", func() {
		err := fs.WriteFileString("/release.yml", `---
name: release
version: version
compile_time_packages:
- name: compile-pkg
  version: compile-version
  fingerprint: compile-fingerprint
  dependencies:
  - dependency
compiled_packages:
- name: compiled-pkg
  version: compiled-version
  fingerprint: compiled-fingerprint
  original_fingerprint: original-fingerprint
  sha1: compiled-sha1
  stemcell: ubuntu-jammy
  dependencies:
  - dependency
`)
		Expect(err).NotTo(HaveOccurred())

		manifest, err := NewManifestFromPath("/release.yml", fs)
		Expect(err).NotTo(HaveOccurred())

		Expect(manifest.CompileTimePackages).To(Equal([]CompileTimePackageRef{{
			Name:         "compile-pkg",
			Version:      "compile-version",
			Fingerprint:  "compile-fingerprint",
			Dependencies: []string{"dependency"},
		}}))
		Expect(manifest.CompiledPkgs[0].OriginalFingerprint).To(Equal("original-fingerprint"))
	})

	It("omits compile-time package metadata and original fingerprints when empty", func() {
		data, err := yaml.Marshal(Manifest{
			Name:         "release",
			Version:      "version",
			CompiledPkgs: []CompiledPackageRef{{Name: "compiled-pkg"}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("compile_time_packages:"))
		Expect(string(data)).NotTo(ContainSubstring("original_fingerprint:"))
	})

	It("marshals compile-time package metadata and original fingerprints when set", func() {
		manifest := Manifest{
			CompileTimePackages: []CompileTimePackageRef{{
				Name:         "compile-pkg",
				Version:      "compile-version",
				Fingerprint:  "compile-fingerprint",
				Dependencies: []string{"dependency"},
			}},
			CompiledPkgs: []CompiledPackageRef{{
				Name:                "compiled-pkg",
				OriginalFingerprint: "original-fingerprint",
			}},
		}

		data, err := yaml.Marshal(manifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("compile_time_packages:"))
		Expect(string(data)).To(ContainSubstring("original_fingerprint: original-fingerprint"))
	})
})
