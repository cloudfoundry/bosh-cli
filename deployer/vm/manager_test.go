package vm_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
)

var _ = Describe("Manager", func() {
	var (
		fakeCloud                  *fakebmcloud.FakeCloud
		manager                    Manager
		logger                     boshlog.Logger
		expectedNetworksSpec       map[string]interface{}
		expectedCloudProperties    map[string]interface{}
		expectedEnv                map[string]interface{}
		deployment                 bmdepl.Deployment
		fakeVMRepo                 *fakebmconfig.FakeVMRepo
		stemcellRepo               bmconfig.StemcellRepo
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
		fakeAgentClientFactory := fakebmagentclient.NewFakeAgentClientFactory()
		fakeAgentClientFactory.CreateAgentClient = fakeAgentClient
		fakeTemplatesSpecGenerator = fakebmas.NewFakeTemplatesSpecGenerator()
		fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()
		fakeVMRepo = fakebmconfig.NewFakeVMRepo()

		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		fakeUUIDGenerator := &fakeuuid.FakeGenerator{}
		stemcellRepo = bmconfig.NewStemcellRepo(configService, fakeUUIDGenerator)

		manager = NewManagerFactory(
			fakeVMRepo,
			stemcellRepo,
			fakeAgentClientFactory,
			fakeApplySpecFactory,
			fakeTemplatesSpecGenerator,
			fs,
			logger,
		).NewManager(fakeCloud, "fake-mbus-url")

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
		deployment = bmdepl.Deployment{
			Name: "fake-deployment",
			Networks: []bmdepl.Network{
				{
					Name: "fake-network-name",
					Type: "dynamic",
				},
			},
			ResourcePools: []bmdepl.ResourcePool{
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
			Jobs: []bmdepl.Job{
				{
					Name: "fake-job",
					Networks: []bmdepl.JobNetwork{
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
			vm, err := manager.Create(stemcell, deployment)
			Expect(err).ToNot(HaveOccurred())
			expectedVM := NewVM(
				"fake-vm-cid",
				fakeVMRepo,
				stemcellRepo,
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
					StemcellCID:     "fake-stemcell-cid",
					CloudProperties: expectedCloudProperties,
					NetworksSpec:    expectedNetworksSpec,
					Env:             expectedEnv,
				},
			))
		})

		It("updates the current vm record", func() {
			_, err := manager.Create(stemcell, deployment)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeVMRepo.UpdateCurrentCID).To(Equal("fake-vm-cid"))
		})

		Context("when creating the vm fails", func() {
			BeforeEach(func() {
				fakeCloud.CreateVMErr = errors.New("fake-create-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(stemcell, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-error"))
			})
		})
	})
})
