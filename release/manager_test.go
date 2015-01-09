package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
)

var _ = Describe("Manager", func() {

	var (
		releaseManager Manager

		releaseA = &fakebmrel.FakeRelease{ReleaseName: "release-a", ReleaseVersion: "version-a"}
		releaseB = &fakebmrel.FakeRelease{ReleaseName: "release-b", ReleaseVersion: "version-b"}
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)

		releaseManager = NewManager(logger)
	})

	Describe("List", func() {
		It("returns all releases that have been added", func() {
			releaseManager.Add(releaseA)
			releaseManager.Add(releaseB)

			Expect(releaseManager.List()).To(Equal([]Release{releaseA, releaseB}))
		})
	})

	Describe("FindByName", func() {
		It("returns false when no releases have been added", func() {
			_, found := releaseManager.FindByName("release-a")
			Expect(found).To(BeFalse())
		})

		Context("when releases have been added", func() {
			It("returns true and the release with the requested name", func() {
				releaseManager.Add(releaseA)
				releaseManager.Add(releaseB)

				releaseAFound, found := releaseManager.FindByName("release-a")
				Expect(found).To(BeTrue())
				Expect(releaseAFound).To(Equal(releaseA))

				releaseBFound, found := releaseManager.FindByName("release-b")
				Expect(found).To(BeTrue())
				Expect(releaseBFound).To(Equal(releaseB))
			})

			It("returns false when the requested release has not been added", func() {
				releaseManager.Add(releaseA)

				_, found := releaseManager.FindByName("release-c")
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("Find", func() {
		It("returns false when no releases have been added", func() {
			_, found := releaseManager.Find("release-a", "version-a")
			Expect(found).To(BeFalse())
		})

		Context("when releases have been added", func() {
			It("returns true and the release with the requested name", func() {
				releaseManager.Add(releaseA)
				releaseManager.Add(releaseB)

				releaseAFound, found := releaseManager.Find("release-a", "version-a")
				Expect(found).To(BeTrue())
				Expect(releaseAFound).To(Equal(releaseA))

				releaseBFound, found := releaseManager.Find("release-b", "version-b")
				Expect(found).To(BeTrue())
				Expect(releaseBFound).To(Equal(releaseB))
			})

			It("returns false when the requested release version has not been added", func() {
				releaseManager.Add(releaseA)

				_, found := releaseManager.Find("release-a", "version-b")
				Expect(found).To(BeFalse())
			})

			It("returns false when the requested release has not been added", func() {
				_, found := releaseManager.Find("release-a", "version-b")
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("DeleteAll", func() {
		It("deletes all added releases", func() {
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
