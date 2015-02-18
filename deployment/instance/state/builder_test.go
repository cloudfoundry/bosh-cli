package state_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_deployment_release "github.com/cloudfoundry/bosh-micro-cli/deployment/release/mocks"
	mock_state_job "github.com/cloudfoundry/bosh-micro-cli/state/job/mocks"
	mock_template "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmstatejob "github.com/cloudfoundry/bosh-micro-cli/state/job"

	fakebmui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("Builder", describeBuilder)

func describeBuilder() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		logger boshlog.Logger

		mockReleaseJobResolver *mock_deployment_release.MockJobResolver
		mockDependencyCompiler *mock_state_job.MockDependencyCompiler
		mockJobListRenderer    *mock_template.MockJobListRenderer
		mockCompressor         *mock_template.MockRenderedJobListCompressor
		mockBlobstore          *mock_blobstore.MockBlobstore

		stateBuilder Builder
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		mockReleaseJobResolver = mock_deployment_release.NewMockJobResolver(mockCtrl)
		mockDependencyCompiler = mock_state_job.NewMockDependencyCompiler(mockCtrl)
		mockJobListRenderer = mock_template.NewMockJobListRenderer(mockCtrl)
		mockCompressor = mock_template.NewMockRenderedJobListCompressor(mockCtrl)
		mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)
	})

	Describe("Build", func() {
		var (
			mockRenderedJobList        *mock_template.MockRenderedJobList
			mockRenderedJobListArchive *mock_template.MockRenderedJobListArchive

			jobName            string
			instanceID         int
			deploymentManifest bmdeplmanifest.Manifest
			fakeStage          *fakebmui.FakeStage

			releasePackageLibyaml *bmrelpkg.Package
			releasePackageRuby    *bmrelpkg.Package
			releasePackageCPI     *bmrelpkg.Package

			expectCompile *gomock.Call
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
						Properties: bmproperty.Map{
							"fake-job-property": "fake-job-property-value",
						},
					},
				},
				Networks: []bmdeplmanifest.Network{
					{
						Name: "fake-network-name",
						IP:   "fake-network-ip",
						Type: "fake-network-type",
						CloudProperties: bmproperty.Map{
							"fake-network-cloud-property": "fake-network-cloud-property-value",
						},
					},
				},
				Properties: bmproperty.Map{
					"fake-job-property": "fake-global-property-value", //overridden by job property value
				},
			}

			fakeStage = fakebmui.NewFakeStage()

			stateBuilder = NewBuilder(
				mockReleaseJobResolver,
				mockDependencyCompiler,
				mockJobListRenderer,
				mockCompressor,
				mockBlobstore,
				logger,
			)

			releasePackageLibyaml = &bmrelpkg.Package{
				Name:         "libyaml",
				Fingerprint:  "fake-package-source-fingerprint-libyaml",
				SHA1:         "fake-package-source-sha1-libyaml",
				Dependencies: []*bmrelpkg.Package{},
				ArchivePath:  "fake-package-archive-path-libyaml", // only required by compiler...
			}
			releasePackageRuby = &bmrelpkg.Package{
				Name:         "ruby",
				Fingerprint:  "fake-package-source-fingerprint-ruby",
				SHA1:         "fake-package-source-sha1-ruby",
				Dependencies: []*bmrelpkg.Package{releasePackageLibyaml},
				ArchivePath:  "fake-package-archive-path-ruby", // only required by compiler...
			}
			releasePackageCPI = &bmrelpkg.Package{
				Name:         "cpi",
				Fingerprint:  "fake-package-source-fingerprint-cpi",
				SHA1:         "fake-package-source-sha1-cpi",
				Dependencies: []*bmrelpkg.Package{releasePackageRuby},
				ArchivePath:  "fake-package-archive-path-cpi", // only required by compiler...
			}
		})

		JustBeforeEach(func() {
			releaseJob := bmreljob.Job{
				Name:        "fake-release-job-name",
				Fingerprint: "fake-release-job-source-fingerprint",
				Packages:    []*bmrelpkg.Package{releasePackageCPI, releasePackageRuby},
			}
			mockReleaseJobResolver.EXPECT().Resolve("fake-release-job-name", "fake-release-name").Return(releaseJob, nil)

			releaseJobs := []bmreljob.Job{releaseJob}
			compiledPackageRefs := []bmstatejob.CompiledPackageRef{
				{
					Name:        "libyaml",
					Version:     "fake-package-source-fingerprint-libyaml",
					BlobstoreID: "fake-package-compiled-archive-blob-id-libyaml",
					SHA1:        "fake-package-compiled-archive-sha1-libyaml",
				},
				{
					Name:        "ruby",
					Version:     "fake-package-source-fingerprint-ruby",
					BlobstoreID: "fake-package-compiled-archive-blob-id-ruby",
					SHA1:        "fake-package-compiled-archive-sha1-ruby",
				},
				{
					Name:        "cpi",
					Version:     "fake-package-source-fingerprint-cpi",
					BlobstoreID: "fake-package-compiled-archive-blob-id-cpi",
					SHA1:        "fake-package-compiled-archive-sha1-cpi",
				},
			}
			expectCompile = mockDependencyCompiler.EXPECT().Compile(releaseJobs, fakeStage).Return(compiledPackageRefs, nil).AnyTimes()

			jobProperties := bmproperty.Map{
				"fake-job-property": "fake-job-property-value",
			}
			globalProperties := bmproperty.Map{
				"fake-job-property": "fake-global-property-value",
			}
			mockJobListRenderer.EXPECT().Render(releaseJobs, jobProperties, globalProperties, "fake-deployment-name").Return(mockRenderedJobList, nil)

			mockRenderedJobList.EXPECT().DeleteSilently()

			mockCompressor.EXPECT().Compress(mockRenderedJobList).Return(mockRenderedJobListArchive, nil)

			mockRenderedJobListArchive.EXPECT().DeleteSilently()

			mockRenderedJobListArchive.EXPECT().Path().Return("fake-rendered-job-list-archive-path")
			mockRenderedJobListArchive.EXPECT().SHA1().Return("fake-rendered-job-list-archive-sha1")
			mockRenderedJobListArchive.EXPECT().Fingerprint().Return("fake-rendered-job-list-fingerprint")

			mockBlobstore.EXPECT().Add("fake-rendered-job-list-archive-path").Return("fake-rendered-job-list-archive-blob-id", nil)
		})

		It("compiles the dependencies of the jobs", func() {
			expectCompile.Times(1)

			_, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())
		})

		It("builds a new instance state with zero-to-many networks", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.NetworkInterfaces()).To(ContainElement(NetworkRef{
				Name: "fake-network-name",
				Interface: bmproperty.Map{
					"ip":   "fake-network-ip",
					"type": "fake-network-type",
					"cloud_properties": bmproperty.Map{
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

		It("prints ui stages for compiling packages and rendering job templates", func() {
			_, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]fakebmui.PerformCall{
				// compile stages not produced by mockDependencyCompiler
				{Name: "Rendering job templates"},
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

		Context("when multiple packages have the same dependency", func() {
			BeforeEach(func() {
				releasePackageRuby.Dependencies = append(releasePackageRuby.Dependencies, releasePackageLibyaml)
			})

			It("does not recompile dependant packages", func() {
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
		})

		It("builds an instance state that can be converted to an ApplySpec", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.ToApplySpec()).To(Equal(bmas.ApplySpec{
				Deployment: "fake-deployment-name",
				Index:      0,
				Networks: map[string]bmproperty.Map{
					"fake-network-name": bmproperty.Map{
						"ip":   "fake-network-ip",
						"type": "fake-network-type",
						"cloud_properties": bmproperty.Map{
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
}
