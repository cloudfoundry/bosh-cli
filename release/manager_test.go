package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"
)

var _ = Describe("Manager", func() {

	var (
		fakeFS               *fakesys.FakeFileSystem
		fakeExtractor        *testfakes.FakeMultiResponseExtractor
		fakeReleaseValidator *fakebmrel.FakeValidator

		deploymentManifestPath string
		releaseManager         Manager

		releaseA = &fakebmrel.FakeRelease{ReleaseName: "release-a"}
		releaseB = &fakebmrel.FakeRelease{ReleaseName: "release-b"}
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseValidator = fakebmrel.NewFakeValidator()
		logger := boshlog.NewLogger(boshlog.LevelNone)

		deploymentManifestPath = "/fake/manifest.yml"
		releaseManager = NewManager(logger)
	})

	Describe("List", func() {
		BeforeEach(func() {
			fakeFS.TempDirDirs = []string{}
		})

		It("returns all releases that have been extracted", func() {
			releaseManager.Add(releaseA)
			releaseManager.Add(releaseB)

			Expect(releaseManager.List()).To(Equal([]Release{releaseA, releaseB}))
		})
	})

	Describe("Find", func() {
		It("returns false when no releases have been extracted", func() {
			_, found := releaseManager.Find("release-a")
			Expect(found).To(BeFalse())
		})

		Context("when releases have been extracted", func() {
			BeforeEach(func() {
				fakeFS.TempDirDirs = []string{}
			})

			It("returns true and the release with the requested name", func() {
				releaseManager.Add(releaseA)
				releaseManager.Add(releaseB)

				releaseAFound, found := releaseManager.Find("release-a")
				Expect(found).To(BeTrue())
				Expect(releaseAFound).To(Equal(releaseA))

				releaseBFound, found := releaseManager.Find("release-b")
				Expect(found).To(BeTrue())
				Expect(releaseBFound).To(Equal(releaseB))
			})

			It("returns false when the requested release has not been extracted", func() {
				releaseManager.Add(releaseA)

				_, found := releaseManager.Find("release-c")
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("DeleteAll", func() {
		BeforeEach(func() {
			fakeFS.TempDirDirs = []string{}
		})

		It("deletes all extracted releases", func() {
			releaseManager.Add(releaseA)
			releaseManager.Add(releaseB)

			err := releaseManager.DeleteAll()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseManager.List()).To(BeEmpty())
			Expect(releaseA.Exists()).To(BeFalse())
			Expect(releaseB.Exists()).To(BeFalse())
		})
	})
})
