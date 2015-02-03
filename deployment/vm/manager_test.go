package vm_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

var _ = Describe("Manager", func() {
	var (
		fakeCloud                 *fakebmcloud.FakeCloud
		manager                   Manager
		logger                    boshlog.Logger
		expectedNetworkInterfaces map[string]bmproperty.Map
		expectedCloudProperties   bmproperty.Map
		expectedEnv               bmproperty.Map
		deploymentManifest        bmdeplmanifest.Manifest
		fakeVMRepo                *fakebmconfig.FakeVMRepo
		stemcellRepo              bmconfig.StemcellRepo
		fakeDiskDeployer          *fakebmvm.FakeDiskDeployer
		fakeAgentClient           *fakebmagentclient.FakeAgentClient
		stemcell                  bmstemcell.CloudStemcell
		fs                        *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeCloud = fakebmcloud.NewFakeCloud()
		fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
		fakeVMRepo = fakebmconfig.NewFakeVMRepo()

		fakeUUIDGenerator := &fakeuuid.FakeGenerator{}
		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, fakeUUIDGenerator, logger)
		stemcellRepo = bmconfig.NewStemcellRepo(configService, fakeUUIDGenerator)

		fakeDiskDeployer = fakebmvm.NewFakeDiskDeployer()

		manager = NewManagerFactory(
			fakeVMRepo,
			stemcellRepo,
			fakeDiskDeployer,
			fakeUUIDGenerator,
			fs,
			logger,
		).NewManager(fakeCloud, fakeAgentClient)

		fakeCloud.CreateVMCID = "fake-vm-cid"
		expectedNetworkInterfaces = map[string]bmproperty.Map{
			"fake-network-name": bmproperty.Map{
				"type":             "dynamic",
				"ip":               "fake-micro-ip",
				"cloud_properties": bmproperty.Map{},
			},
		}
		expectedCloudProperties = bmproperty.Map{
			"fake-cloud-property-key": "fake-cloud-property-value",
		}
		expectedEnv = bmproperty.Map{
			"fake-env-key": "fake-env-value",
		}
		deploymentManifest = bmdeplmanifest.Manifest{
			Name: "fake-deployment",
			Networks: []bmdeplmanifest.Network{
				{
					Name:            "fake-network-name",
					Type:            "dynamic",
					CloudProperties: bmproperty.Map{},
				},
			},
			ResourcePools: []bmdeplmanifest.ResourcePool{
				{
					Name: "fake-resource-pool-name",
					CloudProperties: bmproperty.Map{
						"fake-cloud-property-key": "fake-cloud-property-value",
					},
					Env: bmproperty.Map{
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
				fs,
				logger,
			)
			Expect(vm).To(Equal(expectedVM))

			Expect(fakeCloud.CreateVMInput).To(Equal(
				fakebmcloud.CreateVMInput{
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
