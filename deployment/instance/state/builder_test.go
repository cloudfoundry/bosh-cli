package state_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_instance_state "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state/mocks"
	mock_deployment_release "github.com/cloudfoundry/bosh-micro-cli/deployment/release/mocks"
	mock_template "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakebmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("Builder", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		logger boshlog.Logger

		mockPackageCompiler    *mock_instance_state.MockPackageCompiler
		mockReleaseJobResolver *mock_deployment_release.MockJobResolver
		mockJobListRenderer    *mock_template.MockJobListRenderer
		mockCompressor         *mock_template.MockRenderedJobListCompressor
		mockBlobstore          *mock_blobstore.MockBlobstore

		mockState *mock_instance_state.MockState

		stateBuilder Builder
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		mockPackageCompiler = mock_instance_state.NewMockPackageCompiler(mockCtrl)
		mockReleaseJobResolver = mock_deployment_release.NewMockJobResolver(mockCtrl)
		mockJobListRenderer = mock_template.NewMockJobListRenderer(mockCtrl)
		mockCompressor = mock_template.NewMockRenderedJobListCompressor(mockCtrl)
		mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)

		mockState = mock_instance_state.NewMockState(mockCtrl)
	})

	Describe("Build", func() {
		var (
			mockRenderedJobList        *mock_template.MockRenderedJobList
			mockRenderedJobListArchive *mock_template.MockRenderedJobListArchive

			jobName            string
			instanceID         int
			deploymentManifest bmdeplmanifest.Manifest
			fakeStage          *fakebmeventlog.FakeStage
		)

		BeforeEach(func() {
			mockRenderedJobList = mock_template.NewMockRenderedJobList(mockCtrl)
			mockRenderedJobListArchive = mock_template.NewMockRenderedJobListArchive(mockCtrl)

			jobName = "fake-deployment-job-name"
			instanceID = 0

			deploymentManifest = bmdeplmanifest.Manifest{
				Name: "fake-deployment-name",
				Jobs: []bmdeplmanifest.Job{
					{
						Name: "fake-deployment-job-name",
						Networks: []bmdeplmanifest.JobNetwork{
							{
								Name: "fake-network-name",
							},
						},
						Templates: []bmdeplmanifest.ReleaseJobRef{
							{
								Name:    "fake-release-job-name",
								Release: "fake-release-name",
							},
						},
						RawProperties: map[interface{}]interface{}{
							"fake-job-property": "fake-job-property-value",
						},
					},
				},
				Networks: []bmdeplmanifest.Network{
					{
						Name: "fake-network-name",
						IP:   "fake-network-ip",
						Type: "fake-network-type",
						RawCloudProperties: map[interface{}]interface{}{
							"fake-network-cloud-property": "fake-network-cloud-property-value",
						},
					},
				},
			}

			fakeStage = fakebmeventlog.NewFakeStage()

			stateBuilder = NewBuilder(
				mockPackageCompiler,
				mockReleaseJobResolver,
				mockJobListRenderer,
				mockCompressor,
				mockBlobstore,
				logger,
			)
		})

		JustBeforeEach(func() {
			releasePackageLibyaml := bmrel.Package{
				Name:         "libyaml",
				Fingerprint:  "fake-package-source-fingerprint-libyaml",
				SHA1:         "fake-package-source-sha1-libyaml",
				Dependencies: []*bmrel.Package{},
				ArchivePath:  "fake-package-archive-path-libyaml", // only required by compiler...
			}
			releasePackageRuby := bmrel.Package{
				Name:         "ruby",
				Fingerprint:  "fake-package-source-fingerprint-ruby",
				SHA1:         "fake-package-source-sha1-ruby",
				Dependencies: []*bmrel.Package{&releasePackageLibyaml},
				ArchivePath:  "fake-package-archive-path-ruby", // only required by compiler...
			}
			releasePackageCPI := bmrel.Package{
				Name:         "cpi",
				Fingerprint:  "fake-package-source-fingerprint-cpi",
				SHA1:         "fake-package-source-sha1-cpi",
				Dependencies: []*bmrel.Package{&releasePackageRuby},
				ArchivePath:  "fake-package-archive-path-cpi", // only required by compiler...
			}
			releaseJob := bmrel.Job{
				Name:        "fake-release-job-name",
				Fingerprint: "fake-release-job-source-fingerprint",
				Packages:    []*bmrel.Package{&releasePackageCPI, &releasePackageRuby},
			}
			mockReleaseJobResolver.EXPECT().Resolve("fake-release-job-name", "fake-release-name").Return(releaseJob, nil)

			compiledPackageLibyaml := PackageRef{
				Name:    "libyaml",
				Version: "fake-package-source-fingerprint-libyaml",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-libyaml",
					BlobstoreID: "fake-package-compiled-archive-blob-id-libyaml",
				},
			}
			compiledPackageRuby := PackageRef{
				Name:    "ruby",
				Version: "fake-package-source-fingerprint-ruby",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-ruby",
					BlobstoreID: "fake-package-compiled-archive-blob-id-ruby",
				},
			}
			compiledPackageCPI := PackageRef{
				Name:    "cpi",
				Version: "fake-package-source-fingerprint-cpi",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-cpi",
					BlobstoreID: "fake-package-compiled-archive-blob-id-cpi",
				},
			}

			compiledLibyamlDeps := map[string]PackageRef{}
			compiledRubyDeps := map[string]PackageRef{"libyaml": compiledPackageLibyaml}
			compiledCPIDeps := map[string]PackageRef{"libyaml": compiledPackageLibyaml, "ruby": compiledPackageRuby}
			gomock.InOrder(
				mockPackageCompiler.EXPECT().Compile(&releasePackageLibyaml, compiledLibyamlDeps).Return(compiledPackageLibyaml, nil),
				mockPackageCompiler.EXPECT().Compile(&releasePackageRuby, compiledRubyDeps).Return(compiledPackageRuby, nil),
				mockPackageCompiler.EXPECT().Compile(&releasePackageCPI, compiledCPIDeps).Return(compiledPackageCPI, nil),
			)

			releaseJobs := []bmrel.Job{releaseJob}
			jobProperties := map[string]interface{}{
				"fake-job-property": "fake-job-property-value",
			}
			mockJobListRenderer.EXPECT().Render(releaseJobs, jobProperties, "fake-deployment-name").Return(mockRenderedJobList, nil)

			mockRenderedJobList.EXPECT().DeleteSilently()

			mockCompressor.EXPECT().Compress(mockRenderedJobList).Return(mockRenderedJobListArchive, nil)

			mockRenderedJobListArchive.EXPECT().DeleteSilently()

			mockRenderedJobListArchive.EXPECT().Path().Return("fake-rendered-job-list-archive-path")
			mockRenderedJobListArchive.EXPECT().SHA1().Return("fake-rendered-job-list-archive-sha1")
			mockRenderedJobListArchive.EXPECT().Fingerprint().Return("fake-rendered-job-list-fingerprint")

			mockBlobstore.EXPECT().Add("fake-rendered-job-list-archive-path").Return("fake-rendered-job-list-archive-blob-id", nil)
		})

		It("builds a new instance state with zero-to-many networks", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.NetworkInterfaces()).To(ContainElement(NetworkRef{
				Name: "fake-network-name",
				Interface: map[string]interface{}{
					"ip":   "fake-network-ip",
					"type": "fake-network-type",
					"cloud_properties": map[string]interface{}{
						"fake-network-cloud-property": "fake-network-cloud-property-value",
					},
				},
			}))
			Expect(state.NetworkInterfaces()).To(HaveLen(1))
		})

		It("builds a new instance state with zero-to-many rendered jobs from one or more releases", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.RenderedJobs()).To(ContainElement(JobRef{
				Name:    "fake-release-job-name",
				Version: "fake-release-job-source-fingerprint",
			}))

			// multiple jobs are rendered in a single archive
			Expect(state.RenderedJobListArchive()).To(Equal(BlobRef{
				BlobstoreID: "fake-rendered-job-list-archive-blob-id",
				SHA1:        "fake-rendered-job-list-archive-sha1",
			}))
			Expect(state.RenderedJobs()).To(HaveLen(1))
		})

		It("prints event logs when rendering job templates", func() {
			_, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
				Name: "Rendering job templates",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		It("builds a new instance state with the compiled packages required by the release jobs", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "cpi",
				Version: "fake-package-source-fingerprint-cpi",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-cpi",
					BlobstoreID: "fake-package-compiled-archive-blob-id-cpi",
				},
			}))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "ruby",
				Version: "fake-package-source-fingerprint-ruby",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-ruby",
					BlobstoreID: "fake-package-compiled-archive-blob-id-ruby",
				},
			}))
		})

		It("prints event logs when compiles packages", func() {
			_, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
				Name: "Compiling package 'libyaml/fake-package-source-fingerprint-libyaml'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
			Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
				Name: "Compiling package 'ruby/fake-package-source-fingerprint-ruby'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
			Expect(fakeStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
				Name: "Compiling package 'cpi/fake-package-source-fingerprint-cpi'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		It("builds a new instance state that includes transitively dependent compiled packages", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "cpi",
				Version: "fake-package-source-fingerprint-cpi",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-cpi",
					BlobstoreID: "fake-package-compiled-archive-blob-id-cpi",
				},
			}))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "ruby",
				Version: "fake-package-source-fingerprint-ruby",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-ruby",
					BlobstoreID: "fake-package-compiled-archive-blob-id-ruby",
				},
			}))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "libyaml",
				Version: "fake-package-source-fingerprint-libyaml",
				Archive: BlobRef{
					SHA1:        "fake-package-compiled-archive-sha1-libyaml",
					BlobstoreID: "fake-package-compiled-archive-blob-id-libyaml",
				},
			}))
			Expect(state.CompiledPackages()).To(HaveLen(3))
		})

		It("builds an instance state that can be converted to an ApplySpec", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.ToApplySpec()).To(Equal(bmas.ApplySpec{
				Deployment: "fake-deployment-name",
				Index:      0,
				Networks: map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"ip":   "fake-network-ip",
						"type": "fake-network-type",
						"cloud_properties": map[string]interface{}{
							"fake-network-cloud-property": "fake-network-cloud-property-value",
						},
					},
				},
				Job: bmas.Job{
					Name: "fake-deployment-job-name",
					Templates: []bmas.Blob{
						{
							Name:    "fake-release-job-name",
							Version: "fake-release-job-source-fingerprint",
						},
					},
				},
				Packages: map[string]bmas.Blob{
					"cpi": bmas.Blob{
						Name:        "cpi",
						Version:     "fake-package-source-fingerprint-cpi",
						SHA1:        "fake-package-compiled-archive-sha1-cpi",
						BlobstoreID: "fake-package-compiled-archive-blob-id-cpi",
					},
					"ruby": bmas.Blob{
						Name:        "ruby",
						Version:     "fake-package-source-fingerprint-ruby",
						SHA1:        "fake-package-compiled-archive-sha1-ruby",
						BlobstoreID: "fake-package-compiled-archive-blob-id-ruby",
					},
					"libyaml": bmas.Blob{
						Name:        "libyaml",
						Version:     "fake-package-source-fingerprint-libyaml",
						SHA1:        "fake-package-compiled-archive-sha1-libyaml",
						BlobstoreID: "fake-package-compiled-archive-blob-id-libyaml",
					},
				},
				RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{
					BlobstoreID: "fake-rendered-job-list-archive-blob-id",
					SHA1:        "fake-rendered-job-list-archive-sha1",
				},
				ConfigurationHash: "fake-rendered-job-list-fingerprint",
			}))
		})
	})
})
