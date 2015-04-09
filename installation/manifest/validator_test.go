package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/installation/manifest"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	birel "github.com/cloudfoundry/bosh-init/release"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
	birelset "github.com/cloudfoundry/bosh-init/release/set"

	fakebirel "github.com/cloudfoundry/bosh-init/release/fakes"
)

var _ = Describe("Validator", func() {
	var (
		logger         boshlog.Logger
		releaseManager birel.Manager
		validator      Validator

		releases      []birelmanifest.ReleaseRef
		validManifest Manifest
		fakeRelease   *fakebirel.FakeRelease
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		releaseManager = birel.NewManager(logger)

		releases = []birelmanifest.ReleaseRef{
			{
				Name:    "provided-valid-release-name",
				Version: "1.0",
			},
		}

		validManifest = Manifest{
			Name: "fake-installation-name",
			Template: ReleaseJobRef{
				Name:    "cpi",
				Release: "provided-valid-release-name",
			},
			Properties: biproperty.Map{
				"fake-prop-key": "fake-prop-value",
				"fake-prop-map-key": biproperty.Map{
					"fake-prop-key": "fake-prop-value",
				},
			},
		}

		fakeRelease = fakebirel.New("provided-valid-release-name", "1.0")
		fakeRelease.ReleaseJobs = []bireljob.Job{{Name: "fake-job-name"}}
		releaseManager.Add(fakeRelease)
	})

	JustBeforeEach(func() {
		releaseResolver := birelset.NewResolver(releaseManager, logger)
		err := releaseResolver.Filter(releases)
		Expect(err).ToNot(HaveOccurred())
		validator = NewValidator(logger, releaseResolver)
	})

	Describe("Validate", func() {
		It("does not error if deployment is valid", func() {
			manifest := validManifest

			err := validator.Validate(manifest)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates template must be fully specified", func() {
			manifest := Manifest{}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.name must be provided"))
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.release must be provided"))
		})

		It("validates template.name is not blank", func() {
			manifest := Manifest{
				Template: ReleaseJobRef{
					Name: " ",
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.name must be provided"))
		})

		It("validates template.release is not blank", func() {
			manifest := Manifest{
				Template: ReleaseJobRef{
					Release: " ",
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.release must be provided"))
		})

		It("validates the release is available", func() {
			manifest := Manifest{
				Template: ReleaseJobRef{
					Release: "not-provided-valid-release-name",
				},
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.release 'not-provided-valid-release-name' must refer to a provided release"))
		})
	})
})
