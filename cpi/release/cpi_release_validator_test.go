package release_test

import (
	"code.google.com/p/gomock/gomock"
	cpirel "github.com/cloudfoundry/bosh-init/cpi/release"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biinstallmanifestfakes "github.com/cloudfoundry/bosh-init/installation/manifest/fakes"
	biinstalltarballmocks "github.com/cloudfoundry/bosh-init/installation/tarball/mocks"
	birelfakes "github.com/cloudfoundry/bosh-init/release/fakes"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
	bireleasemocks "github.com/cloudfoundry/bosh-init/release/mocks"
	birelsetman "github.com/cloudfoundry/bosh-init/release/set/manifest"
	birelsetmanfakes "github.com/cloudfoundry/bosh-init/release/set/manifest/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validator", func() {
	var mockCtrl *gomock.Controller
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		releaseSetManifestParser    *birelsetmanfakes.FakeParser
		releaseSetManifestValidator *birelsetmanfakes.FakeValidator
		installationValidator       *biinstallmanifestfakes.FakeValidator
		tarballProvider             *biinstalltarballmocks.MockProvider
		releaseExtractor            *bireleasemocks.MockExtractor
		releaseManager              *bireleasemocks.MockManager
		cpiReleaseValidator         cpirel.CPIReleaseValidator
		installManifest             biinstallmanifest.Manifest
	)

	BeforeEach(func() {
		releaseSetManifestParser = birelsetmanfakes.NewFakeParser()
		releaseSetManifestValidator = birelsetmanfakes.NewFakeValidator()
		installationValidator = biinstallmanifestfakes.NewFakeValidator()

		tarballProvider = biinstalltarballmocks.NewMockProvider(mockCtrl)
		releaseExtractor = bireleasemocks.NewMockExtractor(mockCtrl)
		releaseManager = bireleasemocks.NewMockManager(mockCtrl)

		installManifest = biinstallmanifest.Manifest{
			Template: biinstallmanifest.ReleaseJobRef{
				Release: "some-release-name",
				Name:    "some-job-name",
			},
		}

		cpiReleaseValidator = cpirel.NewCPIReleaseValidator(
			releaseSetManifestParser,
			releaseSetManifestValidator,
			installationValidator,
			tarballProvider,
			releaseExtractor,
			releaseManager,
		)
	})

	It("handles errors parsing the release set manifest", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		releaseSetManifestParser.ParseErr = errors.New("wow that didn't work")

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Parsing release set manifest 'some-path': wow that didn't work"))
	})

	It("handles errors validating the release set manifest", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		releaseSetManifest := birelsetman.Manifest{}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: errors.New("couldn't validate that")},
		})

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Validating release set manifest: couldn't validate that"))
	})

	It("handles installation manifest validation errors", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		releaseSetManifest := birelsetman.Manifest{}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: nil},
		})
		installationValidator.SetValidateBehavior([]biinstallmanifestfakes.ValidateOutput{
			{Err: errors.New("nope")},
		})

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Validating installation manifest: nope"))

	})

	It("errors when the referenced release isn't in the release set manifest", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		cpiReleaseRef := birelmanifest.ReleaseRef{
			Name: "some-other-release-name",
		}
		releaseSetManifest := birelsetman.Manifest{
			Releases: []birelmanifest.ReleaseRef{cpiReleaseRef},
		}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: nil},
		})
		installationValidator.SetValidateBehavior([]biinstallmanifestfakes.ValidateOutput{
			{Err: nil},
		})

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("installation release 'some-release-name' must refer to a release in releases"))
	})

	It("handles errors downloading the release", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		cpiReleaseRef := birelmanifest.ReleaseRef{
			Name: "some-release-name",
		}
		releaseSetManifest := birelsetman.Manifest{
			Releases: []birelmanifest.ReleaseRef{cpiReleaseRef},
		}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: nil},
		})
		installationValidator.SetValidateBehavior([]biinstallmanifestfakes.ValidateOutput{
			{Err: nil},
		})

		tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return("", errors.New("hey, that download failed"))

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("hey, that download failed"))
	})

	It("handles errors extracting the release", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		cpiReleaseRef := birelmanifest.ReleaseRef{
			Name: "some-release-name",
		}
		releaseSetManifest := birelsetman.Manifest{
			Releases: []birelmanifest.ReleaseRef{cpiReleaseRef},
		}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: nil},
		})
		installationValidator.SetValidateBehavior([]biinstallmanifestfakes.ValidateOutput{
			{Err: nil},
		})

		releasePath := "some/release/path"
		tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return(releasePath, nil)

		releaseExtractor.EXPECT().Extract(releasePath).Return(nil, errors.New("boom"))

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Extracting release 'some/release/path': boom"))
	})

	It("validates the release set; validates the installation manifest; downloads, extracts, adds the the manager, and validates the cpi release", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		cpiReleaseRef := birelmanifest.ReleaseRef{
			Name: "some-release-name",
		}
		releaseSetManifest := birelsetman.Manifest{
			Releases: []birelmanifest.ReleaseRef{cpiReleaseRef},
		}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: nil},
		})
		installationValidator.SetValidateBehavior([]biinstallmanifestfakes.ValidateOutput{
			{Err: nil},
		})

		// it should download the release
		releasePath := "some/release/path"
		tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return(releasePath, nil)

		// it should extract the release
		cpiRelease := birelfakes.New("some-release-name", "some-release-version")
		cpiRelease.ReleaseJobs = []bireljob.Job{
			{
				Name:      "some-job-name",
				Templates: map[string]string{"some-template": "bin/cpi"},
			},
		}
		releaseExtractor.EXPECT().Extract(releasePath).Return(cpiRelease, nil)

		// it should add the release the release manager so it can be used?/cleaned up later
		releaseManager.EXPECT().Add(cpiRelease)

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).ToNot(HaveOccurred())

		// it should have validates the release set manifest
		Expect(
			releaseSetManifestValidator.ValidateInputs,
		).To(
			Equal([]birelsetmanfakes.ValidateInput{
				{Manifest: releaseSetManifest},
			}),
		)

		// it should have validated the installation manifest
		Expect(
			installationValidator.ValidateInputs,
		).To(
			Equal([]biinstallmanifestfakes.ValidateInput{
				{
					InstallationManifest: installManifest,
					ReleaseSetManifest:   releaseSetManifest,
				},
			}),
		)

		// it printed a stage
		Expect(stage.PerformCalls).To(Equal([]fakeui.PerformCall{
			{Name: "Validating release 'some-release-name'"},
		}))
	})

	It("validates that the release has the job", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		cpiReleaseRef := birelmanifest.ReleaseRef{
			Name: "some-release-name",
		}
		releaseSetManifest := birelsetman.Manifest{
			Releases: []birelmanifest.ReleaseRef{cpiReleaseRef},
		}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: nil},
		})
		installationValidator.SetValidateBehavior([]biinstallmanifestfakes.ValidateOutput{
			{Err: nil},
		})

		releasePath := "some/release/path"
		tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return(releasePath, nil)

		cpiRelease := birelfakes.New("some-release-name", "some-release-version")
		cpiRelease.ReleaseJobs = []bireljob.Job{
			{Name: "some-other-job-name"},
		}

		releaseExtractor.EXPECT().Extract(releasePath).Return(cpiRelease, nil)

		releaseManager.EXPECT().Add(cpiRelease)

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Invalid CPI release 'some-release-name': CPI release must contain specified job 'some-job-name'"))
	})

	It("validates the release job has a 'bin/cpi' template", func() {
		deploymentManifestPath := "some-path"
		stage := fakeui.NewFakeStage()

		cpiReleaseRef := birelmanifest.ReleaseRef{
			Name: "some-release-name",
		}
		releaseSetManifest := birelsetman.Manifest{
			Releases: []birelmanifest.ReleaseRef{cpiReleaseRef},
		}
		releaseSetManifestParser.ParseManifest = releaseSetManifest
		releaseSetManifestValidator.SetValidateBehavior([]birelsetmanfakes.ValidateOutput{
			{Err: nil},
		})
		installationValidator.SetValidateBehavior([]biinstallmanifestfakes.ValidateOutput{
			{Err: nil},
		})

		releasePath := "some/release/path"
		tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return(releasePath, nil)

		cpiRelease := birelfakes.New("some-release-name", "some-release-version")
		cpiRelease.ReleaseJobs = []bireljob.Job{
			{
				Name:      "some-job-name",
				Templates: map[string]string{"some-template": "bin/not-the-right-binary"},
			},
		}
		releaseExtractor.EXPECT().Extract(releasePath).Return(cpiRelease, nil)

		releaseManager.EXPECT().Add(cpiRelease)

		err := cpiReleaseValidator.RegisterValidCpiReleaseSpecifiedIn(deploymentManifestPath, installManifest, stage)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Invalid CPI release 'some-release-name': Specified CPI release job 'some-job-name' must contain a template that renders to target 'bin/cpi'"))
	})

})
