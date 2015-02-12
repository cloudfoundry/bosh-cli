package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"

	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
)

var _ = Describe("Validator", func() {
	var (
		logger         boshlog.Logger
		releaseManager bmrel.Manager
		validator      Validator

		releases      []bmrelmanifest.ReleaseRef
		validManifest Manifest
		fakeRelease   *fakebmrel.FakeRelease
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		releaseManager = bmrel.NewManager(logger)

		releases = []bmrelmanifest.ReleaseRef{
			{
				Name:    "provided-valid-release-name",
				Version: "1.0",
			},
		}

		validManifest = Manifest{
			Name:    "fake-installation-name",
			Release: "provided-valid-release-name",
			Properties: bmproperty.Map{
				"fake-prop-key": "fake-prop-value",
				"fake-prop-map-key": bmproperty.Map{
					"fake-prop-key": "fake-prop-value",
				},
			},
		}

		fakeRelease = fakebmrel.New("provided-valid-release-name", "1.0")
		fakeRelease.ReleaseJobs = []bmreljob.Job{{Name: "fake-job-name"}}
		releaseManager.Add(fakeRelease)
	})

	JustBeforeEach(func() {
		releaseResolver := bmrelset.NewResolver(releaseManager, logger)
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

		It("validates release is not blank", func() {
			manifest := Manifest{
				Release: " ",
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.release must be provided"))
		})

		It("validates the release is available", func() {
			manifest := Manifest{
				Release: "not-provided-valid-release-name",
			}

			err := validator.Validate(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.release 'not-provided-valid-release-name' must refer to a provided release"))
		})
	})
})
