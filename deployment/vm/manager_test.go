package vm_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebicloud "github.com/cloudfoundry/bosh-init/cloud/fakes"
	fakebiconfig "github.com/cloudfoundry/bosh-init/config/fakes"
	fakebiagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient/fakes"
	fakebivm "github.com/cloudfoundry/bosh-init/deployment/vm/fakes"

	. "github.com/cloudfoundry/bosh-init/deployment/vm"
)

var _ = Describe("Manager", func() {
	var (
		fakeCloud                 *fakebicloud.FakeCloud
		manager                   Manager
		logger                    boshlog.Logger
		expectedNetworkInterfaces map[string]biproperty.Map
		expectedCloudProperties   biproperty.Map
		expectedEnv               biproperty.Map
		deploymentManifest        bideplmanifest.Manifest
		fakeVMRepo                *fakebiconfig.FakeVMRepo
		stemcellRepo              biconfig.StemcellRepo
		fakeDiskDeployer          *fakebivm.FakeDiskDeployer
		fakeAgentClient           *fakebiagentclient.FakeAgentClient
		stemcell                  bistemcell.CloudStemcell
		fs                        *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeCloud = fakebicloud.NewFakeCloud()
		fakeAgentClient = fakebiagentclient.NewFakeAgentClient()
		fakeVMRepo = fakebiconfig.NewFakeVMRepo()

		fakeUUIDGenerator := &fakeuuid.FakeGenerator{}
		configService := biconfig.NewFileSystemDeploymentConfigService(fs, fakeUUIDGenerator, logger)
		configService.SetConfigPath("/fake/path")
		stemcellRepo = biconfig.NewStemcellRepo(configService, fakeUUIDGenerator)

		fakeDiskDeployer = fakebivm.NewFakeDiskDeployer()

		manager = NewManagerFactory(
			fakeVMRepo,
			stemcellRepo,
			fakeDiskDeployer,
			fakeUUIDGenerator,
			fs,
			logger,
		).NewManager(fakeCloud, fakeAgentClient)

		fakeCloud.CreateVMCID = "fake-vm-cid"
		expectedNetworkInterfaces = map[string]biproperty.Map{
			"fake-network-name": biproperty.Map{
				"type":             "dynamic",
				"ip":               "fake-micro-ip",
				"cloud_properties": biproperty.Map{},
			},
		}
		expectedCloudProperties = biproperty.Map{
			"fake-cloud-property-key": "fake-cloud-property-value",
		}
		expectedEnv = biproperty.Map{
			"fake-env-key": "fake-env-value",
		}
		deploymentManifest = bideplmanifest.Manifest{
			Name: "fake-deployment",
			Networks: []bideplmanifest.Network{
				{
					Name:            "fake-network-name",
					Type:            "dynamic",
					CloudProperties: biproperty.Map{},
				},
			},
			ResourcePools: []bideplmanifest.ResourcePool{
				{
					Name: "fake-resource-pool-name",
					CloudProperties: biproperty.Map{
						"fake-cloud-property-key": "fake-cloud-property-value",
					},
					Env: biproperty.Map{
						"fake-env-key": "fake-env-value",
					},
				},
			},
			Jobs: []bideplmanifest.Job{
				{
					Name: "fake-job",
					Networks: []bideplmanifest.JobNetwork{
						{
							Name:      "fake-network-name",
							StaticIPs: []string{"fake-micro-ip"},
						},
					},
				},
			},
		}

		stemcellRecord := biconfig.StemcellRecord{CID: "fake-stemcell-cid"}
		stemcell = bistemcell.NewCloudStemcell(stemcellRecord, stemcellRepo, fakeCloud)
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
				fs,
				logger,
			)
			Expect(vm).To(Equal(expectedVM))

			Expect(fakeCloud.CreateVMInput).To(Equal(
				fakebicloud.CreateVMInput{
					AgentID:            "fake-uuid-0",
					StemcellCID:        "fake-stemcell-cid",
					CloudProperties:    expectedCloudProperties,
					NetworksInterfaces: expectedNetworkInterfaces,
					Env:                expectedEnv,
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
