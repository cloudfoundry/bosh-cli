package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"

	fakebirel "github.com/cloudfoundry/bosh-init/release/fakes"

	. "github.com/cloudfoundry/bosh-init/release/set/manifest"
)

var _ = Describe("Validator", func() {
	var (
		logger    boshlog.Logger
		validator Validator

		validManifest Manifest
		fakeRelease   *fakebirel.FakeRelease
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		validManifest = Manifest{
			Releases: []birelmanifest.ReleaseRef{
				{
					Name:    "fake-release-name",
					Version: "1.0",
					URL:     "file://fake-release-path",
				},
			},
		}

		fakeRelease = fakebirel.New("fake-release-name", "1.0")
		fakeRelease.ReleaseJobs = []bireljob.Job{{Name: "fake-job-name"}}
	})

	JustBeforeEach(func() {
		validator = NewValidator(logger)
	})

	Describe("Validate", func() {
		It("does not error if deployment is valid", func() {
			manifest := validManifest

			err := validator.Validate(manifest)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates there is at least one release", func() {
			manifest := Manifest{
				Releases: []birelmanifest.ReleaseRef{},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases must contain at least 1 release"))
		})

		It("validates releases have names", func() {
			manifest := Manifest{
				Releases: []birelmanifest.ReleaseRef{{}},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0].name must be provided"))
		})

		It("validates releases have urls", func() {
			manifest := Manifest{
				Releases: []birelmanifest.ReleaseRef{
					{Name: "fake-release-name"},
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0].url must be provided"))
		})

		It("validates releases have valid urls", func() {
			manifest := Manifest{
				Releases: []birelmanifest.ReleaseRef{
					{Name: "fake-release-name", URL: "invalid-url"},
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0].url must be a valid file URL (file://)"))
		})

		It("validates releases are unique", func() {
			manifest := Manifest{
				Releases: []birelmanifest.ReleaseRef{
					{Name: "fake-release-name"},
					{Name: "fake-release-name"},
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[1].name 'fake-release-name' must be unique"))
		})

		It("validates release version is a SemVer", func() {
			manifest := Manifest{
				Releases: []birelmanifest.ReleaseRef{
					{Name: "fake-release-name", Version: "not-a-semver"},
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("releases[0].version 'not-a-semver' must be a semantic version (name: 'fake-release-name')"))
		})

		It("allows release versions to be 'latest'", func() {
			manifest := validManifest
			manifest.Releases = []birelmanifest.ReleaseRef{
				{Name: "fake-release-name", Version: "latest", URL: "file://fake-release-path"},
			}

			err := validator.Validate(manifest)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
