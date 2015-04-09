package manifest_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"

	. "github.com/cloudfoundry/bosh-init/release/set/manifest"
)

var _ = Describe("Parser", func() {
	var (
		comboManifestPath string
		fakeFs            *fakesys.FakeFileSystem
		parser            Parser
	)

	BeforeEach(func() {
		comboManifestPath = "fake-deployment-path"
		fakeFs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		parser = NewParser(fakeFs, logger)
	})

	Context("when combo manifest path does not exist", func() {
		BeforeEach(func() {
			err := fakeFs.RemoveAll(comboManifestPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when parser fails to read the combo manifest file", func() {
		BeforeEach(func() {
			fakeFs.ReadFileError = errors.New("fake-read-file-error")
		})

		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	BeforeEach(func() {
		contents := `
---
releases:
- name: fake-release-name-1
  version: fake-release-version-1
- name: fake-release-name-2
  version: fake-release-version-2
name: unknown-keys-are-ignored
`
		fakeFs.WriteFileString(comboManifestPath, contents)
	})

	It("parses release set manifest from combo manifest file", func() {
		deploymentManifest, err := parser.Parse(comboManifestPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(deploymentManifest).To(Equal(Manifest{
			Releases: []bmrelmanifest.ReleaseRef{
				{
					Name:    "fake-release-name-1",
					Version: "fake-release-version-1",
				},
				{
					Name:    "fake-release-name-2",
					Version: "fake-release-version-2",
				},
			},
		}))
	})
})
