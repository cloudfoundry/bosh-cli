package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/release"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fake_release "github.com/cloudfoundry/bosh-init/release/fakes"
)

var _ = Describe("Manager", func() {

	var (
		releaseManager Manager

		releaseA = fake_release.New("release-a", "version-a")
		releaseB = fake_release.New("release-b", "version-b")
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
				Expect(releaseAFound).To(Equal([]Release{releaseA}))

				releaseBFound, found := releaseManager.FindByName("release-b")
				Expect(found).To(BeTrue())
				Expect(releaseBFound).To(Equal([]Release{releaseB}))
			})

			Context("when multiple versions of the same release have been added", func() {
				It("returns true and the release with the requested name", func() {
					releaseA10 := fake_release.New("release-a", "1.0")
					releaseA11 := fake_release.New("release-a", "1.1")
					releaseB10 := fake_release.New("release-b", "1.0")
					releaseManager.Add(releaseA10)
					releaseManager.Add(releaseA11)
					releaseManager.Add(releaseB10)

					releaseAFound, found := releaseManager.FindByName("release-a")
					Expect(found).To(BeTrue())
					Expect(releaseAFound).To(Equal([]Release{releaseA10, releaseA11}))
				})
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
