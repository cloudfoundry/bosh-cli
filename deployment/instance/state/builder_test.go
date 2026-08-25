package state_test

import (
	biac "github.com/cloudfoundry/bosh-agent/v2/agentclient"
	bias "github.com/cloudfoundry/bosh-agent/v2/agentclient/applyspec"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/blobstore/blobstorefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/release/releasefakes"
	boshjob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	bistatejob "github.com/cloudfoundry/bosh-cli/v7/state/job"
	"github.com/cloudfoundry/bosh-cli/v7/state/job/jobfakes"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/templatescompilerfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("Builder", func() {
	var (
		logger boshlog.Logger

		mockReleaseJobResolver *releasefakes.FakeJobResolver
		mockDependencyCompiler *jobfakes.FakeDependencyCompiler
		mockJobListRenderer    *templatescompilerfakes.FakeJobListRenderer
		mockCompressor         *templatescompilerfakes.FakeRenderedJobListCompressor
		mockBlobstore          *blobstorefakes.FakeBlobstore

		stateBuilder Builder
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		mockReleaseJobResolver = &releasefakes.FakeJobResolver{}
		mockDependencyCompiler = &jobfakes.FakeDependencyCompiler{}
		mockJobListRenderer = &templatescompilerfakes.FakeJobListRenderer{}
		mockCompressor = &templatescompilerfakes.FakeRenderedJobListCompressor{}
		mockBlobstore = &blobstorefakes.FakeBlobstore{}
	})

	Describe("BuildInitialState", func() {
		var (
			jobName            string
			instanceID         int
			deploymentManifest bideplmanifest.Manifest
		)

		BeforeEach(func() {
			jobName = "fake-deployment-job-name"
			instanceID = 0

			deploymentManifest = bideplmanifest.Manifest{
				Name: "fake-deployment-name",
				Jobs: []bideplmanifest.Job{
					{
						Name: "fake-deployment-job-name",
						Networks: []bideplmanifest.JobNetwork{
							{
								Name:      "fake-network-name",
								StaticIPs: []string{"1.2.3.4"},
							},
						},
						Templates: []bideplmanifest.ReleaseJobRef{
							{
								Name:    "job-name",
								Release: "fake-release-name",
								Properties: &biproperty.Map{
									"fake-template-property": "fake-template-property-value",
								},
							},
						},
						Properties: biproperty.Map{
							"fake-job-property": "fake-job-property-value",
						},
					},
				},
				Networks: []bideplmanifest.Network{
					{
						Name: "fake-network-name",
						Type: "fake-network-type",
						CloudProperties: biproperty.Map{
							"fake-network-cloud-property": "fake-network-cloud-property-value",
						},
					},
				},
				Properties: biproperty.Map{
					"fake-job-property": "fake-global-property-value", // overridden by job property value
				},
			}

			stateBuilder = NewBuilder(
				mockReleaseJobResolver,
				mockDependencyCompiler,
				mockJobListRenderer,
				mockCompressor,
				mockBlobstore,
				logger,
			)
		})

		It("generates an initial apply spec", func() {
			state, err := stateBuilder.BuildInitialState(jobName, instanceID, deploymentManifest)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.ToApplySpec()).To(Equal(bias.ApplySpec{
				Name:             "fake-deployment-job-name",
				NodeID:           "0",
				AvailabilityZone: "unknown",
				Deployment:       "fake-deployment-name",
				Index:            0,
				Job: bias.Job{
					Name:      "fake-deployment-job-name",
					Templates: []bias.Blob{},
				},
				Packages: map[string]bias.Blob{},
				Networks: map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"type":    "fake-network-type",
						"default": []bideplmanifest.NetworkDefault{"dns", "gateway"},
						"ip":      "1.2.3.4",
						"cloud_properties": biproperty.Map{
							"fake-network-cloud-property": "fake-network-cloud-property-value",
						},
					},
				},
				ConfigurationHash: "unused-configuration-hash",
			}))
		})
	})

	Describe("Build", func() {
		var (
			mockRenderedJobList        *templatescompilerfakes.FakeRenderedJobList
			mockRenderedJobListArchive *templatescompilerfakes.FakeRenderedJobListArchive

			jobName            string
			instanceID         int
			deploymentManifest bideplmanifest.Manifest
			fakeStage          *testui.Stage

			agentState            biac.AgentState
			expectedIP            string
			releasePackageLibyaml *boshpkg.Package
			releasePackageRuby    *boshpkg.Package
			releasePackageCPI     *boshpkg.Package
		)

		BeforeEach(func() {
			mockRenderedJobList = &templatescompilerfakes.FakeRenderedJobList{}
			mockRenderedJobListArchive = &templatescompilerfakes.FakeRenderedJobListArchive{}

			jobName = "fake-deployment-job-name"
			instanceID = 0
			expectedIP = "1.2.3.4"

			deploymentManifest = bideplmanifest.Manifest{
				Name: "fake-deployment-name",
				Jobs: []bideplmanifest.Job{
					{
						Name: "fake-deployment-job-name",
						Networks: []bideplmanifest.JobNetwork{
							{
								Name:      "fake-network-name",
								StaticIPs: []string{"1.2.3.4"},
							},
						},
						Templates: []bideplmanifest.ReleaseJobRef{
							{
								Name:    "job-name",
								Release: "fake-release-name",
								Properties: &biproperty.Map{
									"fake-template-property": "fake-template-property-value",
								},
							},
						},
						Properties: biproperty.Map{
							"fake-job-property": "fake-job-property-value",
						},
					},
				},
				Networks: []bideplmanifest.Network{
					{
						Name: "fake-network-name",
						Type: "fake-network-type",
						CloudProperties: biproperty.Map{
							"fake-network-cloud-property": "fake-network-cloud-property-value",
						},
					},
				},
				Properties: biproperty.Map{
					"fake-job-property": "fake-global-property-value", // overridden by job property value
				},
			}

			agentState = biac.AgentState{
				NetworkSpecs: map[string]biac.NetworkSpec{
					"fake-network-name": {
						IP: "1.2.3.5",
					},
				},
			}

			fakeStage = &testui.Stage{}

			stateBuilder = NewBuilder(
				mockReleaseJobResolver,
				mockDependencyCompiler,
				mockJobListRenderer,
				mockCompressor,
				mockBlobstore,
				logger,
			)

			releasePackageLibyaml = boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"libyaml", "libyaml-fp", "libyaml-path", "libyaml-sha1"), nil)

			releasePackageRuby = boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"ruby", "ruby-fp", "ruby-path", "ruby-sha1"), []string{"libyaml"})
			err := releasePackageRuby.AttachDependencies([]*boshpkg.Package{releasePackageLibyaml})
			Expect(err).ToNot(HaveOccurred())

			releasePackageCPI = boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"cpi", "cpi-fp", "cpi-path", "cpi-sha1"), []string{"ruby"})
			err = releasePackageCPI.AttachDependencies([]*boshpkg.Package{releasePackageRuby})
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if mockJobListRenderer.RenderCallCount() > 0 {
				_, _, _, _, _, address := mockJobListRenderer.RenderArgsForCall(0)
				Expect(address).To(Equal(expectedIP))
			}
		})

		JustBeforeEach(func() {
			releaseJob := *boshjob.NewJob(NewResource("job-name", "job-fp", nil))
			err := releaseJob.AttachPackages([]*boshpkg.Package{releasePackageCPI, releasePackageRuby})
			Expect(err).ToNot(HaveOccurred())

			mockReleaseJobResolver.ResolveReturns(releaseJob, nil)

			compiledPackageRefs := []bistatejob.CompiledPackageRef{
				{
					Name:        "libyaml",
					Version:     "libyaml-fp",
					BlobstoreID: "libyaml-blob-id",
					SHA1:        "libyaml-sha1",
				},
				{
					Name:        "ruby",
					Version:     "ruby-fp",
					BlobstoreID: "ruby-blob-id",
					SHA1:        "ruby-sha1",
				},
				{
					Name:        "cpi",
					Version:     "cpi-fp",
					BlobstoreID: "cpi-bosh-id",
					SHA1:        "cpi-sha1",
				},
			}
			mockDependencyCompiler.CompileReturns(compiledPackageRefs, nil)

			mockJobListRenderer.RenderReturns(mockRenderedJobList, nil)

			mockCompressor.CompressReturns(mockRenderedJobListArchive, nil)

			mockRenderedJobListArchive.PathReturns("fake-rendered-job-list-archive-path")
			mockRenderedJobListArchive.SHA1Returns("fake-rendered-job-list-archive-sha1")

			mockBlobstore.AddReturns("fake-rendered-job-list-archive-blob-id", nil)
		})

		It("compiles the dependencies of the jobs", func() {
			_, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockDependencyCompiler.CompileCallCount()).To(Equal(1))
			Expect(mockRenderedJobList.DeleteSilentlyCallCount()).To(Equal(1))
			Expect(mockRenderedJobListArchive.DeleteSilentlyCallCount()).To(Equal(1))
		})

		It("builds a new instance state with zero-to-many networks", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.NetworkInterfaces()).To(ContainElement(NetworkRef{
				Name: "fake-network-name",
				Interface: map[string]interface{}{
					"type":    "fake-network-type",
					"default": []bideplmanifest.NetworkDefault{"dns", "gateway"},
					"ip":      "1.2.3.4",
					"cloud_properties": biproperty.Map{
						"fake-network-cloud-property": "fake-network-cloud-property-value",
					},
				},
			}))
			Expect(state.NetworkInterfaces()).To(HaveLen(1))
		})

		Context("dynamic network without IP address", func() {
			Context("single network", func() {
				BeforeEach(func() {
					deploymentManifest.Jobs[0].Networks[0].StaticIPs = nil
					deploymentManifest.Networks[0].Type = "dynamic"
					expectedIP = "1.2.3.5"
				})

				It("should not fail", func() {
					state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
					Expect(err).ToNot(HaveOccurred())

					Expect(state.NetworkInterfaces()).To(ContainElement(NetworkRef{
						Name: "fake-network-name",
						Interface: map[string]interface{}{
							"type":    "dynamic",
							"default": []bideplmanifest.NetworkDefault{"dns", "gateway"},
							"cloud_properties": biproperty.Map{
								"fake-network-cloud-property": "fake-network-cloud-property-value",
							},
						},
					}))
					Expect(state.NetworkInterfaces()).To(HaveLen(1))
				})
			})

			Context("multiple networks", func() {
				BeforeEach(func() {
					expectedIP = "1.2.3.6"
					deploymentManifest.Networks = append(
						deploymentManifest.Networks,
						bideplmanifest.Network{
							Name: "fake-dynamic-network-name",
							Type: "dynamic",
							CloudProperties: biproperty.Map{
								"fake-network-cloud-property": "fake-network-cloud-property-value",
							},
						},
					)
					deploymentManifest.Jobs[0].Networks = append(
						deploymentManifest.Jobs[0].Networks,
						bideplmanifest.JobNetwork{
							Name:     "fake-dynamic-network-name",
							Defaults: []bideplmanifest.NetworkDefault{bideplmanifest.NetworkDefaultDNS, bideplmanifest.NetworkDefaultGateway},
						},
					)
					agentState.NetworkSpecs["fake-dynamic-network-name"] = biac.NetworkSpec{
						IP: "1.2.3.6",
					}
				})

				It("should not fail", func() {
					state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
					Expect(err).ToNot(HaveOccurred())

					Expect(state.NetworkInterfaces()).To(ContainElement(NetworkRef{
						Name: "fake-dynamic-network-name",
						Interface: map[string]interface{}{
							"type":    "dynamic",
							"default": []bideplmanifest.NetworkDefault{"dns", "gateway"},
							"cloud_properties": biproperty.Map{
								"fake-network-cloud-property": "fake-network-cloud-property-value",
							},
						},
					}))
					Expect(state.NetworkInterfaces()).To(HaveLen(2))
				})
			})
		})

		It("builds a new instance state with zero-to-many rendered jobs from one or more releases", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.RenderedJobs()).To(ContainElement(JobRef{
				Name:    "job-name",
				Version: "job-fp",
			}))

			// multiple jobs are rendered in a single archive
			Expect(state.RenderedJobListArchive()).To(Equal(BlobRef{
				BlobstoreID: "fake-rendered-job-list-archive-blob-id",
				SHA1:        "fake-rendered-job-list-archive-sha1",
			}))
			Expect(state.RenderedJobs()).To(HaveLen(1))
		})

		It("prints ui stages for compiling packages and rendering job templates", func() {
			_, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
				// compile stages not produced by mockDependencyCompiler
				{Name: "Rendering job templates"},
			}))
		})

		It("builds a new instance state with the compiled packages required by the release jobs", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "cpi",
				Version: "cpi-fp",
				Archive: BlobRef{
					SHA1:        "cpi-sha1",
					BlobstoreID: "cpi-bosh-id",
				},
			}))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "ruby",
				Version: "ruby-fp",
				Archive: BlobRef{
					SHA1:        "ruby-sha1",
					BlobstoreID: "ruby-blob-id",
				},
			}))
		})

		It("builds a new instance state that includes transitively dependent compiled packages", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "cpi",
				Version: "cpi-fp",
				Archive: BlobRef{
					SHA1:        "cpi-sha1",
					BlobstoreID: "cpi-bosh-id",
				},
			}))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "ruby",
				Version: "ruby-fp",
				Archive: BlobRef{
					SHA1:        "ruby-sha1",
					BlobstoreID: "ruby-blob-id",
				},
			}))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "libyaml",
				Version: "libyaml-fp",
				Archive: BlobRef{
					SHA1:        "libyaml-sha1",
					BlobstoreID: "libyaml-blob-id",
				},
			}))
			Expect(state.CompiledPackages()).To(HaveLen(3))
		})

		Context("when multiple packages have the same dependency", func() {
			BeforeEach(func() {
				releasePackageRuby.Dependencies = append(releasePackageRuby.Dependencies, releasePackageLibyaml)
			})

			It("does not recompile dependant packages", func() {
				state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
				Expect(err).ToNot(HaveOccurred())

				Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
					Name:    "cpi",
					Version: "cpi-fp",
					Archive: BlobRef{
						SHA1:        "cpi-sha1",
						BlobstoreID: "cpi-bosh-id",
					},
				}))
				Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
					Name:    "ruby",
					Version: "ruby-fp",
					Archive: BlobRef{
						SHA1:        "ruby-sha1",
						BlobstoreID: "ruby-blob-id",
					},
				}))
				Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
					Name:    "libyaml",
					Version: "libyaml-fp",
					Archive: BlobRef{
						SHA1:        "libyaml-sha1",
						BlobstoreID: "libyaml-blob-id",
					},
				}))
				Expect(state.CompiledPackages()).To(HaveLen(3))
			})
		})

		It("builds an instance state that can be converted to an ApplySpec", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, fakeStage, agentState)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.ToApplySpec()).To(Equal(bias.ApplySpec{
				Name:             "fake-deployment-job-name",
				NodeID:           "0",
				AvailabilityZone: "unknown",
				Deployment:       "fake-deployment-name",
				Index:            0,
				Networks: map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"type":    "fake-network-type",
						"default": []bideplmanifest.NetworkDefault{"dns", "gateway"},
						"ip":      "1.2.3.4",
						"cloud_properties": biproperty.Map{
							"fake-network-cloud-property": "fake-network-cloud-property-value",
						},
					},
				},
				Job: bias.Job{
					Name: "fake-deployment-job-name",
					Templates: []bias.Blob{
						{
							Name:    "job-name",
							Version: "job-fp",
						},
					},
				},
				Packages: map[string]bias.Blob{
					"cpi": {
						Name:        "cpi",
						Version:     "cpi-fp",
						SHA1:        "cpi-sha1",
						BlobstoreID: "cpi-bosh-id",
					},
					"ruby": {
						Name:        "ruby",
						Version:     "ruby-fp",
						SHA1:        "ruby-sha1",
						BlobstoreID: "ruby-blob-id",
					},
					"libyaml": {
						Name:        "libyaml",
						Version:     "libyaml-fp",
						SHA1:        "libyaml-sha1",
						BlobstoreID: "libyaml-blob-id",
					},
				},
				RenderedTemplatesArchive: bias.RenderedTemplatesArchiveSpec{
					BlobstoreID: "fake-rendered-job-list-archive-blob-id",
					SHA1:        "fake-rendered-job-list-archive-sha1",
				},
				ConfigurationHash: "unused-configuration-hash",
			}))
		})
	})
})
