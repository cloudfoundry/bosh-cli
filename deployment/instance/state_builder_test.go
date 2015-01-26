package instance_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_instance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/mocks"
	mock_release_job "github.com/cloudfoundry/bosh-micro-cli/release/job/mocks"
	mock_template "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakeboshuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
)

var _ = Describe("StateBuilder", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		logger boshlog.Logger

		mockReleaseJobResolver *mock_release_job.MockResolver
		mockJobListRenderer    *mock_template.MockJobListRenderer
		mockCompressor         *mock_template.MockRenderedJobListCompressor
		mockBlobstore          *mock_blobstore.MockBlobstore

		fakeUUIDGenerator *fakeboshuuid.FakeGenerator

		mockState *mock_instance.MockState

		stateBuilder StateBuilder
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		mockReleaseJobResolver = mock_release_job.NewMockResolver(mockCtrl)
		mockJobListRenderer = mock_template.NewMockJobListRenderer(mockCtrl)
		mockCompressor = mock_template.NewMockRenderedJobListCompressor(mockCtrl)
		mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)

		fakeUUIDGenerator = fakeboshuuid.NewFakeGenerator()

		mockState = mock_instance.NewMockState(mockCtrl)

	})

	Describe("Build", func() {
		var (
			mockRenderedJobList        *mock_template.MockRenderedJobList
			mockRenderedJobListArchive *mock_template.MockRenderedJobListArchive

			jobName            string
			instanceID         int
			deploymentManifest bmdeplmanifest.Manifest
			stemcellApplySpec  bmstemcell.ApplySpec
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

			stemcellApplySpec = bmstemcell.ApplySpec{
				Packages: map[string]bmstemcell.Blob{
					"cpi": bmstemcell.Blob{
						Name:        "cpi",
						Version:     "fake-fingerprint-cpi",
						SHA1:        "fake-sha1-cpi",
						BlobstoreID: "fake-package-blob-id-cpi",
					},
					"ruby": bmstemcell.Blob{
						Name:        "ruby",
						Version:     "fake-fingerprint-ruby",
						SHA1:        "fake-sha1-ruby",
						BlobstoreID: "fake-package-blob-id-ruby",
					},
				},
			}

			stateBuilder = NewStateBuilder(
				mockReleaseJobResolver,
				mockJobListRenderer,
				mockCompressor,
				mockBlobstore,
				fakeUUIDGenerator,
				logger,
			)
		})

		JustBeforeEach(func() {
			releaseJobRefs := []bmreljob.Reference{
				{
					Name:    "fake-release-job-name",
					Release: "fake-release-name",
				},
			}
			releaseJobs := []bmrel.Job{
				{
					Name:        "fake-release-job-name",
					Fingerprint: "fake-release-job-fingerprint",
				},
			}
			mockReleaseJobResolver.EXPECT().ResolveEach(releaseJobRefs).Return(releaseJobs, nil)

			jobProperties := map[string]interface{}{
				"fake-job-property": "fake-job-property-value",
			}
			mockJobListRenderer.EXPECT().Render(releaseJobs, jobProperties, "fake-deployment-name").Return(mockRenderedJobList, nil)

			mockRenderedJobList.EXPECT().DeleteSilently()

			mockCompressor.EXPECT().Compress(mockRenderedJobList).Return(mockRenderedJobListArchive, nil)

			mockRenderedJobListArchive.EXPECT().DeleteSilently()

			fakeUUIDGenerator.GeneratedUuid = "fake-rendered-job-list-archive-blob-id"

			mockRenderedJobListArchive.EXPECT().Path().Return("fake-rendered-job-list-archive-path")
			mockRenderedJobListArchive.EXPECT().SHA1().Return("fake-rendered-job-list-archive-sha1")
			mockRenderedJobListArchive.EXPECT().Fingerprint().Return("fake-rendered-job-list-fingerprint")

			mockBlobstore.EXPECT().Save("fake-rendered-job-list-archive-path", "fake-rendered-job-list-archive-blob-id")
		})

		It("builds a new instance state with zero-to-many networks", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, stemcellApplySpec)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.NetworkInterfaces()).To(HaveLen(1))
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
		})

		It("builds a new instance state with zero-to-many rendered jobs from one or more releases", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, stemcellApplySpec)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.RenderedJobs()).To(HaveLen(1))
			Expect(state.RenderedJobs()).To(ContainElement(JobRef{
				Name:    "fake-release-job-name",
				Version: "fake-release-job-fingerprint",
			}))

			// multiple jobs are rendered in a single archive
			Expect(state.RenderedJobListArchive()).To(Equal(BlobRef{
				BlobstoreID: "fake-rendered-job-list-archive-blob-id",
				SHA1:        "fake-rendered-job-list-archive-sha1",
			}))
		})

		It("builds a new instance state with zero-to-many compiled packages from one or more releases", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, stemcellApplySpec)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.CompiledPackages()).To(HaveLen(2))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "cpi",
				Version: "fake-fingerprint-cpi",
				Archive: BlobRef{
					SHA1:        "fake-sha1-cpi",
					BlobstoreID: "fake-package-blob-id-cpi",
				},
			}))
			Expect(state.CompiledPackages()).To(ContainElement(PackageRef{
				Name:    "ruby",
				Version: "fake-fingerprint-ruby",
				Archive: BlobRef{
					SHA1:        "fake-sha1-ruby",
					BlobstoreID: "fake-package-blob-id-ruby",
				},
			}))
		})

		It("builds an instance state that can be converted to an ApplySpec", func() {
			state, err := stateBuilder.Build(jobName, instanceID, deploymentManifest, stemcellApplySpec)
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
							Version: "fake-release-job-fingerprint",
						},
					},
				},
				Packages: map[string]bmas.Blob{
					"cpi": bmas.Blob{
						Name:        "cpi",
						Version:     "fake-fingerprint-cpi",
						SHA1:        "fake-sha1-cpi",
						BlobstoreID: "fake-package-blob-id-cpi",
					},
					"ruby": bmas.Blob{
						Name:        "ruby",
						Version:     "fake-fingerprint-ruby",
						SHA1:        "fake-sha1-ruby",
						BlobstoreID: "fake-package-blob-id-ruby",
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
