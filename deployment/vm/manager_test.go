package vm_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

var _ = Describe("Manager", func() {
	var (
		fakeCloud                  *fakebmcloud.FakeCloud
		manager                    Manager
		logger                     boshlog.Logger
		expectedNetworksSpec       map[string]interface{}
		expectedCloudProperties    map[string]interface{}
		expectedEnv                map[string]interface{}
		deploymentManifest         bmdeplmanifest.Manifest
		fakeVMRepo                 *fakebmconfig.FakeVMRepo
		stemcellRepo               bmconfig.StemcellRepo
		fakeDiskDeployer           *fakebmvm.FakeDiskDeployer
		fakeAgentClient            *fakebmagentclient.FakeAgentClient
		fakeTemplatesSpecGenerator *fakebmas.FakeTemplatesSpecGenerator
		fakeApplySpecFactory       *fakebmas.FakeApplySpecFactory
		stemcell                   bmstemcell.CloudStemcell
		fs                         *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeCloud = fakebmcloud.NewFakeCloud()
		fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
		fakeTemplatesSpecGenerator = fakebmas.NewFakeTemplatesSpecGenerator()
		fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()
		fakeVMRepo = fakebmconfig.NewFakeVMRepo()

		fakeUUIDGenerator := &fakeuuid.FakeGenerator{}
		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, fakeUUIDGenerator, logger)
		stemcellRepo = bmconfig.NewStemcellRepo(configService, fakeUUIDGenerator)

		fakeDiskDeployer = fakebmvm.NewFakeDiskDeployer()

		manager = NewManagerFactory(
			fakeVMRepo,
			stemcellRepo,
			fakeDiskDeployer,
			fakeApplySpecFactory,
			fakeTemplatesSpecGenerator,
			fakeUUIDGenerator,
			fs,
			logger,
		).NewManager(fakeCloud, fakeAgentClient, "fake-mbus-url")

		fakeCloud.CreateVMCID = "fake-vm-cid"
		expectedNetworksSpec = map[string]interface{}{
			"fake-network-name": map[string]interface{}{
				"type":             "dynamic",
				"ip":               "fake-micro-ip",
				"cloud_properties": map[string]interface{}{},
			},
		}
		expectedCloudProperties = map[string]interface{}{
			"fake-cloud-property-key": "fake-cloud-property-value",
		}
		expectedEnv = map[string]interface{}{
			"fake-env-key": "fake-env-value",
		}
		deploymentManifest = bmdeplmanifest.Manifest{
			Name: "fake-deployment",
			Networks: []bmdeplmanifest.Network{
				{
					Name: "fake-network-name",
					Type: "dynamic",
				},
			},
			ResourcePools: []bmdeplmanifest.ResourcePool{
				{
					Name: "fake-resource-pool-name",
					RawCloudProperties: map[interface{}]interface{}{
						"fake-cloud-property-key": "fake-cloud-property-value",
					},
					RawEnv: map[interface{}]interface{}{
						"fake-env-key": "fake-env-value",
					},
				},
			},
			Jobs: []bmdeplmanifest.Job{
				{
					Name: "fake-job",
					Networks: []bmdeplmanifest.JobNetwork{
						{
							Name:      "fake-network-name",
							StaticIPs: []string{"fake-micro-ip"},
						},
					},
				},
			},
		}

		stemcellRecord := bmconfig.StemcellRecord{CID: "fake-stemcell-cid"}
		stemcell = bmstemcell.NewCloudStemcell(stemcellRecord, stemcellRepo, fakeCloud)
	})

	Describe("Create", func() {
		It("creates a VM", func() {
			vm, err := manager.Create(stemcell, deploymentManifest)
			Expect(err).ToNot(HaveOccurred())
			expectedVM := NewVM(
				"fake-vm-cid",
				fakeVMRepo,
				stemcellRepo,
				fakeDiskDeployer,
				fakeAgentClient,
				fakeCloud,
				fakeTemplatesSpecGenerator,
				fakeApplySpecFactory,
				"fake-mbus-url",
				fs,
				logger,
			)
			Expect(vm).To(Equal(expectedVM))

			Expect(fakeCloud.CreateVMInput).To(Equal(
				fakebmcloud.CreateVMInput{
					AgentID:         "fake-uuid-0",
					StemcellCID:     "fake-stemcell-cid",
					CloudProperties: expectedCloudProperties,
					NetworksSpec:    expectedNetworksSpec,
					Env:             expectedEnv,
				},
			))
		})

		It("updates the current vm record", func() {
			_, err := manager.Create(stemcell, deploymentManifest)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeVMRepo.UpdateCurrentCID).To(Equal("fake-vm-cid"))
		})

		Context("when creating the vm fails", func() {
			BeforeEach(func() {
				fakeCloud.CreateVMErr = errors.New("fake-create-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(stemcell, deploymentManifest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-error"))
			})
		})
	})
})
