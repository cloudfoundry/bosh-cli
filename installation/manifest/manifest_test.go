package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	birelsetman "github.com/cloudfoundry/bosh-init/release/set/manifest"

	"errors"
	"github.com/cloudfoundry/bosh-init/installation/manifest"
	"github.com/cloudfoundry/bosh-init/installation/manifest/fakes"
)

var _ = Describe("ParseAndValidateFrom", func() {
	var (
		deploymentManifestPath = "some-manifest-path"
		parser                 *fakes.FakeParser
		validator              *fakes.FakeValidator
		releaseSetManifest     birelsetman.Manifest
	)
	BeforeEach(func() {
		releaseSetManifest = birelsetman.Manifest{}
		parser = fakes.NewFakeParser()
		validator = fakes.NewFakeValidator()
	})

	It("returns the installation manifest", func() {
		expectedManifest := manifest.Manifest{}
		parser.ParseManifest = expectedManifest
		validator.SetValidateBehavior([]fakes.ValidateOutput{
			{Err: nil},
		})

		installManifest, err := manifest.ParseAndValidateFrom(deploymentManifestPath, parser, validator, releaseSetManifest)
		Expect(err).ToNot(HaveOccurred())

		Expect(installManifest).To(Equal(expectedManifest))
	})

	It("handles errors parsing the installation manifest", func() {
		parser.ParseErr = errors.New("wow that didn't work")

		_, err := manifest.ParseAndValidateFrom(deploymentManifestPath, parser, validator, releaseSetManifest)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Parsing installation manifest 'some-manifest-path': wow that didn't work"))
	})

	It("handles installation manifest validation errors", func() {
		parser.ParseManifest = manifest.Manifest{}
		validator.SetValidateBehavior([]fakes.ValidateOutput{
			{Err: errors.New("nope")},
		})

		_, err := manifest.ParseAndValidateFrom(deploymentManifestPath, parser, validator, releaseSetManifest)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Validating installation manifest: nope"))

	})
})
