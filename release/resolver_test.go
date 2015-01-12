package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"

	"github.com/cloudfoundry/bosh-agent/logger"

	fake_release "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
)

var _ = Describe("Resolver", func() {

	var (
		myLogger        logger.Logger
		releaseManager  Manager
		releaseVersions []bmrelmanifest.ReleaseRef

		releaseA10 = fake_release.New("release-a", "1.0")
	)

	BeforeEach(func() {
		logger := logger.NewLogger(logger.LevelNone)

		releaseVersions = []bmrelmanifest.ReleaseRef{}
		releaseManager = NewManager(logger)
	})

	createResolver := func() Resolver {
		return NewResolver(myLogger, releaseManager, releaseVersions)
	}
	addReleaseVersionRule := func(name, version string) {
		releaseVersions = append(releaseVersions, bmrelmanifest.ReleaseRef{Name: name, Version: version})
	}

	Context("when a release version has not been specified", func() {
		Context("when no release is available with that name", func() {
			It("returns an error", func() {
				_, err := createResolver().Find("release-a")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Release 'release-a' is not available"))
			})
		})

		It("resolves releases using the latest available version", func() {
			releaseManager.Add(releaseA10)
			release, err := createResolver().Find("release-a")
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
				release, err := createResolver().Find("release-a")
				Expect(err).ToNot(HaveOccurred())
				Expect(release.Version()).To(Equal("1.2"))

				releaseManager.Add(fake_release.New("release-a", "1.3-alpha"))
				release, err = createResolver().Find("release-a")
				Expect(err).ToNot(HaveOccurred())
				Expect(release.Version()).To(Equal("1.3-alpha"))
			})
		})
	})

	Context("when a release version has been specified", func() {
		BeforeEach(func() {
			addReleaseVersionRule("release-a", "1.1")
		})

		Context("when no release is available with that name", func() {
			It("returns an error", func() {
				releaseManager.Add(fake_release.New("release-a", "1.0"))

				_, err := createResolver().Find("release-a")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No version of 'release-a' matches '1.1"))
			})
		})

		Context("and a matching version of the release is available", func() {
			BeforeEach(func() {
				releaseManager.Add(fake_release.New("release-a", "1.0"))
				releaseManager.Add(fake_release.New("release-a", "1.1"))
				releaseManager.Add(fake_release.New("release-a", "1.2"))
			})

			It("returns that release", func() {
				release, err := createResolver().Find("release-a")
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
			addReleaseVersionRule("release-a", "latest")
		})

		It("returns the release with the greatest version", func() {
			release, err := createResolver().Find("release-a")
			Expect(err).ToNot(HaveOccurred())
			Expect(release.Version()).To(Equal("1.2"))
		})
	})

	It("reports problems with version constraints", func() {
		releaseManager.Add(fake_release.New("release-a", "1.0"))
		addReleaseVersionRule("release-a", "_")
		_, err := createResolver().Find("release-a")
		Expect(err.Error()).To(ContainSubstring("Parsing requested version for 'release-a': Malformed constraint: _"))
	})

	It("reports problems with versions", func() {
		releaseManager.Add(fake_release.New("release-a", "_"))
		_, err := createResolver().Find("release-a")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Parsing version of 'release-a': Malformed version: _"))
	})

	It("reports problems if a release version is specified more than once", func() {
		releaseManager.Add(fake_release.New("release-a", "_"))
		addReleaseVersionRule("release-a", "1.0")
		addReleaseVersionRule("release-a", "1.1")
		_, err := createResolver().Find("release-a")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Duplicate release 'release-a'"))
	})
})
