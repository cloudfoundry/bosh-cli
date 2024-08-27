package manifest_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	birelmanifest "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
)

var _ = Describe("Validator", func() {
	var (
		logger             boshlog.Logger
		releaseSetManifest birelsetmanifest.Manifest
		validator          Validator

		releases      []birelmanifest.ReleaseRef
		validManifest Manifest
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		releases = []birelmanifest.ReleaseRef{
			{Name: "provided-valid-release-name"},
		}

		validManifest = Manifest{
			Name: "fake-installation-name",
			Templates: []ReleaseJobRef{
				{Name: "cpi", Release: "provided-valid-release-name"},
			},
			Properties: biproperty.Map{
				"fake-prop-key": "fake-prop-value",
				"fake-prop-map-key": biproperty.Map{
					"fake-prop-key": "fake-prop-value",
				},
			},
		}

		releaseSetManifest = birelsetmanifest.Manifest{
			Releases: releases,
		}

		validator = NewValidator(logger)
	})

	Describe("Validate", func() {
		It("does not error if deployment is valid", func() {
			manifest := validManifest

			err := validator.Validate(manifest, releaseSetManifest)
			Expect(err).ToNot(HaveOccurred())
		})

		It("errors when validating an empty manifest", func() {
			manifest := Manifest{}

			err := validator.Validate(manifest, releaseSetManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("either cloud_provider.templates or cloud_provider.template must be provided and must contain at least one release"))
		})

		It("validates template must be fully specified", func() {
			manifest := Manifest{
				Templates: []ReleaseJobRef{
					{Name: "", Release: ""},
				},
			}

			err := validator.Validate(manifest, releaseSetManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.name must be provided"))
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.release must be provided"))
		})

		It("validates template.name is not blank", func() {
			manifest := Manifest{
				Templates: []ReleaseJobRef{
					{Name: " "},
				},
			}

			err := validator.Validate(manifest, releaseSetManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.name must be provided"))
		})

		It("validates template.release is not blank", func() {
			manifest := Manifest{
				Templates: []ReleaseJobRef{
					{Release: " "},
				},
			}

			err := validator.Validate(manifest, releaseSetManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.release must be provided"))
		})

		It("validates the release is available", func() {
			manifest := Manifest{
				Templates: []ReleaseJobRef{
					{Release: "not-provided-valid-release-name"},
				},
			}

			err := validator.Validate(manifest, releaseSetManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cloud_provider.template.release 'not-provided-valid-release-name' must refer to a release in releases"))
		})

		It("validates the release successfully when multiple valid templates are specified", func() {
			validManifest.Templates = append(validManifest.Templates, ReleaseJobRef{Name: "plugin", Release: "provided-valid-release-name"})

			err := validator.Validate(validManifest, releaseSetManifest)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
