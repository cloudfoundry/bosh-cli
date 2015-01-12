package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"

	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"
)

var _ = Describe("Validator", func() {
	var (
		releaseManager bmrel.Manager
		validator      Validator

		validManifest Manifest
		fakeRelease   *fakebmrel.FakeRelease
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		releaseManager = bmrel.NewManager(logger)
		validator = NewValidator(logger, releaseManager)

		validManifest = Manifest{
			Releases: []bmrelmanifest.ReleaseRef{
				{
					Name:    "fake-release-name",
					Version: "1.0",
				},
			},
		}

		fakeRelease = fakebmrel.New("fake-release-name", "1.0")
		fakeRelease.ReleaseJobs = []bmrel.Job{{Name: "fake-job-name"}}
		releaseManager.Add(fakeRelease)
	})

	Describe("Validate", func() {
		It("does not error if deployment is valid", func() {
			manifest := validManifest

			err := validator.Validate(manifest)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates releases have names", func() {
			manifest := Manifest{
				Releases: []bmrelmanifest.ReleaseRef{{}},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0].name must be provided"))
		})

		It("validates releases have versions", func() {
			manifest := Manifest{
				Releases: []bmrelmanifest.ReleaseRef{{}},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0].version must be provided"))
		})

		It("validates release version is a SemVer", func() {
			manifest := Manifest{
				Releases: []bmrelmanifest.ReleaseRef{
					{Name: "fake-release-name", Version: "not-a-semver"},
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0].version must be a semantic version"))
		})

		It("validates release is available", func() {
			manifest := validManifest
			manifest.Releases = []bmrelmanifest.ReleaseRef{
				{Name: "fake-other-release-name", Version: "1.0"},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0] must refer to an available release"))
		})

		It("allows release versions to be 'latest'", func() {
			manifest := validManifest
			manifest.Releases = []bmrelmanifest.ReleaseRef{
				{Name: "fake-release-name", Version: "latest"},
			}
			releaseManager.Add(fakebmrel.New("fake-release-name", "1.0"))

			err := validator.Validate(manifest)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates releases are unique", func() {
			manifest := Manifest{
				Releases: []bmrelmanifest.ReleaseRef{
					{Name: "fake-release-name"},
					{Name: "fake-release-name"},
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[1].name 'fake-release-name' must be unique"))
		})
	})
})
