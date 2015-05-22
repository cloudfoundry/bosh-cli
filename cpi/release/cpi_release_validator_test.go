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
	biui "github.com/cloudfoundry/bosh-init/ui"
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

	Describe("GetFrom", func() {
		var (
			releaseSetManifestParser    *birelsetmanfakes.FakeParser
			releaseSetManifestValidator *birelsetmanfakes.FakeValidator
			installationParser          *biinstallmanifestfakes.FakeParser
			installationValidator       *biinstallmanifestfakes.FakeValidator
			tarballProvider             *biinstalltarballmocks.MockProvider
			releaseExtractor            *bireleasemocks.MockExtractor
			releaseManager              *bireleasemocks.MockManager
			validatedCpiReleaseSpec     cpirel.ValidatedCpiReleaseSpec
			installManifest             biinstallmanifest.Manifest
			deploymentManifestPath      string
		)

		BeforeEach(func() {
			deploymentManifestPath = "some-path"
			releaseSetManifestParser = birelsetmanfakes.NewFakeParser()
			releaseSetManifestValidator = birelsetmanfakes.NewFakeValidator()
			installationParser = biinstallmanifestfakes.NewFakeParser()
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
			installationParser.ParseManifest = installManifest

			validatedCpiReleaseSpec = cpirel.NewValidatedCpiReleaseSpec(releaseSetManifestParser, installationParser)
		})

		It("parses and validates all the things and returns a release ref", func() {
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

			actualInstallManifest, releaseRef, err := validatedCpiReleaseSpec.GetFrom(deploymentManifestPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseRef).To(Equal(cpiReleaseRef))
			Expect(actualInstallManifest).To(Equal(installManifest))
		})

		It("handles errors parsing the release set manifest", func() {
			releaseSetManifestParser.ParseErr = errors.New("wow that didn't work")

			_, _, err := validatedCpiReleaseSpec.GetFrom(deploymentManifestPath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Parsing release set manifest 'some-path': wow that didn't work"))
		})

		It("handles errors parsing the installation manifest", func() {
			installationParser.ParseErr = errors.New("wow that didn't work")

			_, _, err := validatedCpiReleaseSpec.GetFrom(deploymentManifestPath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Parsing installation manifest 'some-path': wow that didn't work"))
		})

		It("errors when the referenced release isn't in the release set manifest", func() {
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

			_, _, err := validatedCpiReleaseSpec.GetFrom(deploymentManifestPath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("installation release 'some-release-name' must refer to a release in releases"))
		})
	})

	Describe("DownloadAndRegister", func() {
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
				tarballProvider,
				releaseExtractor,
				releaseManager,
			)
		})

		It("handles errors downloading the release", func() {
			stage := fakeui.NewFakeStage()

			cpiReleaseRef := birelmanifest.ReleaseRef{
				Name: "some-release-name",
			}

			tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return("", errors.New("hey, that download failed"))

			err := cpiReleaseValidator.DownloadAndRegister(cpiReleaseRef, installManifest, stage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("hey, that download failed"))
		})

		It("handles errors extracting the release", func() {
			stage := fakeui.NewFakeStage()

			cpiReleaseRef := birelmanifest.ReleaseRef{
				Name: "some-release-name",
			}

			releasePath := "some/release/path"
			tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return(releasePath, nil)

			releaseExtractor.EXPECT().Extract(releasePath).Return(nil, errors.New("boom"))

			err := cpiReleaseValidator.DownloadAndRegister(cpiReleaseRef, installManifest, stage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Extracting release 'some/release/path': boom"))
		})

		It("downloads, extracts, adds the the manager, and validates the cpi release", func() {
			stage := fakeui.NewFakeStage()

			cpiReleaseRef := birelmanifest.ReleaseRef{
				Name: "some-release-name",
			}

			// it should download the release
			releasePath := "some/release/path"
			tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Do(func(releaseRef birelmanifest.ReleaseRef, stage biui.Stage) {
				stage.Perform("I'm the download step", func() error {
					return nil
				})
			}).Return(releasePath, nil)

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

			err := cpiReleaseValidator.DownloadAndRegister(cpiReleaseRef, installManifest, stage)
			Expect(err).ToNot(HaveOccurred())

			// it printed a stage
			Expect(stage.PerformCalls).To(Equal([]*fakeui.PerformCall{
				{Name: "I'm the download step"},
				{Name: "Validating release 'some-release-name'"},
			}))
		})

		It("validates that the release has the job", func() {
			stage := fakeui.NewFakeStage()

			cpiReleaseRef := birelmanifest.ReleaseRef{
				Name: "some-release-name",
			}

			releasePath := "some/release/path"
			tarballProvider.EXPECT().Get(cpiReleaseRef, stage).Return(releasePath, nil)

			cpiRelease := birelfakes.New("some-release-name", "some-release-version")
			cpiRelease.ReleaseJobs = []bireljob.Job{
				{Name: "some-other-job-name"},
			}

			releaseExtractor.EXPECT().Extract(releasePath).Return(cpiRelease, nil)

			releaseManager.EXPECT().Add(cpiRelease)

			err := cpiReleaseValidator.DownloadAndRegister(cpiReleaseRef, installManifest, stage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Invalid CPI release 'some-release-name': CPI release must contain specified job 'some-job-name'"))
		})

		It("validates the release job has a 'bin/cpi' template", func() {
			stage := fakeui.NewFakeStage()

			cpiReleaseRef := birelmanifest.ReleaseRef{
				Name: "some-release-name",
			}

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

			err := cpiReleaseValidator.DownloadAndRegister(cpiReleaseRef, installManifest, stage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Invalid CPI release 'some-release-name': Specified CPI release job 'some-job-name' must contain a template that renders to target 'bin/cpi'"))
		})

	})
})
