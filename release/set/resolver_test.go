package set_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/release/set"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	birel "github.com/cloudfoundry/bosh-init/release"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"

	fake_release "github.com/cloudfoundry/bosh-init/release/fakes"
)

var _ = Describe("Resolver", func() {

	var (
		logger         boshlog.Logger
		releaseManager birel.Manager
		releases       []birelmanifest.ReleaseRef

		releaseA10 = fake_release.New("release-a", "1.0")
		resolver   Resolver
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		releases = []birelmanifest.ReleaseRef{}
		releaseManager = birel.NewManager(logger)
	})

	JustBeforeEach(func() {
		resolver = NewResolver(releaseManager, logger)
	})

	Context("when a release version has not been specified", func() {
		Context("when no release is available with that name", func() {
			It("returns an error", func() {
				_, err := resolver.Find("release-a")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Release 'release-a' is not available"))
			})
		})

		It("resolves releases using the latest available version", func() {
			releaseManager.Add(releaseA10)
			release, err := resolver.Find("release-a")
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(releaseA10))
		})

		Context("when many versions of the same release are available", func() {
			BeforeEach(func() {
				releaseManager.Add(fake_release.New("release-a", "1.2"))
				releaseManager.Add(fake_release.New("release-a", "1.2-alpha"))
				releaseManager.Add(fake_release.New("release-a", "0.9"))
			})

			It("resolves releases using the latest available version", func() {
				release, err := resolver.Find("release-a")
				Expect(err).ToNot(HaveOccurred())
				Expect(release.Version()).To(Equal("1.2"))

				releaseManager.Add(fake_release.New("release-a", "1.3-alpha"))
				release, err = resolver.Find("release-a")
				Expect(err).ToNot(HaveOccurred())
				Expect(release.Version()).To(Equal("1.3-alpha"))
			})
		})
	})

	Context("when a release version has been specified", func() {
		JustBeforeEach(func() {
			resolver.Filter([]birelmanifest.ReleaseRef{
				{Name: "release-a", Version: "1.1"},
			})
		})

		Context("when no release is available with the specified version", func() {
			BeforeEach(func() {
				releaseManager.Add(fake_release.New("release-a", "1.0"))
			})

			It("returns an error", func() {
				_, err := resolver.Find("release-a")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No version of 'release-a' matches '1.1'"))
			})
		})

		Context("and a matching version of the release is available", func() {
			BeforeEach(func() {
				releaseManager.Add(fake_release.New("release-a", "1.0"))
				releaseManager.Add(fake_release.New("release-a", "1.1"))
				releaseManager.Add(fake_release.New("release-a", "1.2"))
			})

			It("returns that release", func() {
				release, err := resolver.Find("release-a")
				Expect(err).ToNot(HaveOccurred())
				Expect(release.Version()).To(Equal("1.1"))
			})
		})
	})

	Context("when a release version is specified as 'latest'", func() {
		BeforeEach(func() {
			releaseManager.Add(fake_release.New("release-a", "1.0"))
			releaseManager.Add(fake_release.New("release-a", "1.2"))
			releaseManager.Add(fake_release.New("release-a", "1.1"))
			resolver.Filter([]birelmanifest.ReleaseRef{
				{Name: "release-a", Version: "latest"},
			})
		})

		It("returns the release with the greatest version", func() {
			release, err := resolver.Find("release-a")
			Expect(err).ToNot(HaveOccurred())
			Expect(release.Version()).To(Equal("1.2"))
		})
	})

	It("reports problems with parsing manifest versions", func() {
		releaseManager.Add(fake_release.New("release-a", "1.0"))
		resolver.Filter([]birelmanifest.ReleaseRef{
			{Name: "release-a", Version: "_"},
		})
		_, err := resolver.Find("release-a")
		Expect(err.Error()).To(ContainSubstring("Parsing version '_' of release 'release-a' from manifest: Malformed constraint: _"))
	})

	It("reports problems with parsing managed release versions", func() {
		releaseManager.Add(fake_release.New("release-a", "_"))
		resolver.Filter([]birelmanifest.ReleaseRef{
			{Name: "release-a", Version: "1.0"},
		})
		_, err := resolver.Find("release-a")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Parsing version '_' of release 'release-a': Malformed version: _"))
	})

	It("reports problems if a release name is specified more than once", func() {
		err := resolver.Filter([]birelmanifest.ReleaseRef{
			{Name: "release-a", Version: "1.0"},
			{Name: "release-a", Version: "1.1"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Duplicate release 'release-a'"))
	})
})
