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
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
)

var _ = Describe("Manager", func() {
	Describe("CreateVM", func() {
		var (
			fakeCloud                  *fakebmcloud.FakeCloud
			manager                    Manager
			logger                     boshlog.Logger
			expectedNetworksSpec       map[string]interface{}
			expectedCloudProperties    map[string]interface{}
			expectedEnv                map[string]interface{}
			deployment                 bmdepl.Deployment
			configService              bmconfig.DeploymentConfigService
			fakeAgentClient            *fakebmagentclient.FakeAgentClient
			fakeTemplatesSpecGenerator *fakebmas.FakeTemplatesSpecGenerator
			fakeApplySpecFactory       *fakebmas.FakeApplySpecFactory
			stemcell                   bmstemcell.CloudStemcell
			fs                         *fakesys.FakeFileSystem
		)

		BeforeEach(func() {
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fs = fakesys.NewFakeFileSystem()
			configService = bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
			fakeCloud = fakebmcloud.NewFakeCloud()
			fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
			fakeAgentClientFactory := fakebmagentclient.NewFakeAgentClientFactory()
			fakeAgentClientFactory.CreateAgentClient = fakeAgentClient
			fakeTemplatesSpecGenerator = fakebmas.NewFakeTemplatesSpecGenerator()
			fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()
			manager = NewManagerFactory(
				fakeAgentClientFactory,
				configService,
				fakeApplySpecFactory,
				fakeTemplatesSpecGenerator,
				fs,
				logger,
			).NewManager(fakeCloud)
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

			stemcell = bmstemcell.CloudStemcell{CID: "fake-stemcell-cid"}
		})

		It("creates a VM", func() {
			vm, err := manager.Create(stemcell, deployment, "fake-mbus-url")
			Expect(err).ToNot(HaveOccurred())
			expectedVM := NewVM(
				"fake-vm-cid",
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

		It("saves the vm record using the config service", func() {
			_, err := manager.Create(stemcell, deployment, "fake-mbus-url")
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := bmconfig.DeploymentConfig{
				VMCID: "fake-vm-cid",
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})

		Context("when creating the vm fails", func() {
			BeforeEach(func() {
				fakeCloud.CreateVMErr = errors.New("fake-create-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(stemcell, deployment, "fake-mbus-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-error"))
			})
		})
	})
})
