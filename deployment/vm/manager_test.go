package vm_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/clock"
	fakebiagentclient "github.com/cloudfoundry/bosh-agent/v2/agentclient/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	fakebicloud "github.com/cloudfoundry/bosh-cli/v7/cloud/fakes"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	fakebiconfig "github.com/cloudfoundry/bosh-cli/v7/config/fakes"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	. "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	fakebivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm/fakes"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
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
		diskCIDs                  []string
		fs                        *fakesys.FakeFileSystem
		fakeTimeService           Clock
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeCloud = fakebicloud.NewFakeCloud()
		fakeAgentClient = &fakebiagentclient.FakeAgentClient{}
		fakeVMRepo = fakebiconfig.NewFakeVMRepo()

		fakeUUIDGenerator := &fakeuuid.FakeGenerator{}
		deploymentStateService := biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, "/fake/path")
		stemcellRepo = biconfig.NewStemcellRepo(deploymentStateService, fakeUUIDGenerator)

		fakeDiskDeployer = fakebivm.NewFakeDiskDeployer()
		fakeTime := time.Date(2016, time.November, 10, 23, 0, 0, 0, time.UTC)
		fakeTimeService = &FakeClock{Times: []time.Time{fakeTime, time.Now().Add(10 * time.Minute)}}
		diskCIDs = []string{"fake-disk-cid-1"}

		manager = NewManager(
			fakeVMRepo,
			stemcellRepo,
			fakeDiskDeployer,
			fakeAgentClient,
			fakeCloud,
			fakeUUIDGenerator,
			fs,
			logger,
			fakeTimeService,
		)

		fakeCloud.CreateVMCID = "fake-vm-cid"
		expectedNetworkInterfaces = map[string]biproperty.Map{
			"fake-network-name": biproperty.Map{
				"type":             "dynamic",
				"ip":               "fake-ip",
				"cloud_properties": biproperty.Map{},
				"default":          []bideplmanifest.NetworkDefault{"dns", "gateway"},
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
							StaticIPs: []string{"fake-ip"},
						},
					},
					ResourcePool: "fake-resource-pool-name",
				},
			},
		}

		stemcellRecord := biconfig.StemcellRecord{CID: "fake-stemcell-cid"}
		stemcell = bistemcell.NewCloudStemcell(stemcellRecord, stemcellRepo, fakeCloud)
	})

	Describe("Create", func() {
		It("creates a VM", func() {
			vm, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
			Expect(err).ToNot(HaveOccurred())
			expectedVM := NewVMWithMetadata(
				"fake-vm-cid",
				fakeVMRepo,
				stemcellRepo,
				fakeDiskDeployer,
				fakeAgentClient,
				fakeCloud,
				clock.NewClock(),
				fs,
				logger,
				bicloud.VMMetadata{
					"deployment":     "fake-deployment",
					"job":            "fake-job",
					"instance_group": "fake-job",
					"index":          "0",
					"director":       "bosh-init",
					"name":           "fake-job/0",
					"created_at":     "2016-11-10T23:00:00Z",
				},
			)
			Expect(vm).To(Equal(expectedVM))

			Expect(fakeCloud.CreateVMInput).To(Equal(
				fakebicloud.CreateVMInput{
					AgentID:            "fake-uuid-0",
					StemcellCID:        "fake-stemcell-cid",
					CloudProperties:    expectedCloudProperties,
					DiskCIDs:           diskCIDs,
					NetworksInterfaces: expectedNetworkInterfaces,
					Env:                expectedEnv,
				},
			))
		})

		It("sets the vm metadata", func() {
			_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeCloud.SetVMMetadataCid).To(Equal("fake-vm-cid"))
			Expect(fakeCloud.SetVMMetadataMetadata).To(Equal(bicloud.VMMetadata{
				"deployment":     "fake-deployment",
				"job":            "fake-job",
				"instance_group": "fake-job",
				"index":          "0",
				"director":       "bosh-init",
				"name":           "fake-job/0",
				"created_at":     "2016-11-10T23:00:00Z",
			}))
		})

		Context("deployment-configured tags", func() {
			It("sets additional tags on vms", func() {
				deploymentManifest.Tags = map[string]string{
					"empty1": "",
					"key1":   "value1",
				}

				expectedEnv = biproperty.Map{
					"fake-env-key": "fake-env-value",
					"bosh": biproperty.Map{
						"tags": map[string]string{
							"empty1": "",
							"key1":   "value1",
						},
					},
				}

				_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeCloud.CreateVMInput).To(Equal(
					fakebicloud.CreateVMInput{
						AgentID:            "fake-uuid-0",
						StemcellCID:        "fake-stemcell-cid",
						CloudProperties:    expectedCloudProperties,
						DiskCIDs:           diskCIDs,
						NetworksInterfaces: expectedNetworkInterfaces,
						Env:                expectedEnv,
					},
				))

				Expect(fakeCloud.SetVMMetadataMetadata).To(Equal(bicloud.VMMetadata{
					"deployment":     "fake-deployment",
					"job":            "fake-job",
					"name":           "fake-job/0",
					"instance_group": "fake-job",
					"index":          "0",
					"director":       "bosh-init",
					"empty1":         "",
					"key1":           "value1",
					"created_at":     "2016-11-10T23:00:00Z",
				}))
			})

			Context("overriding built-in metadata", func() {
				It("gives precedence to deployment tags", func() {
					deploymentManifest.Tags = map[string]string{
						"deployment":     "manifest-deployment",
						"job":            "manifest-job",
						"instance_group": "manifest-instance-group",
						"index":          "7",
						"director":       "manifest-director",
						"name":           "awesome-name",
					}

					_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeCloud.SetVMMetadataMetadata).To(Equal(bicloud.VMMetadata{
						"deployment":     "manifest-deployment",
						"job":            "manifest-job",
						"name":           "awesome-name",
						"instance_group": "manifest-instance-group",
						"index":          "7",
						"director":       "manifest-director",
						"created_at":     "2016-11-10T23:00:00Z",
					}))
				})
			})
		})

		It("updates the current vm record", func() {
			_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeVMRepo.UpdateCurrentCID).To(Equal("fake-vm-cid"))
		})

		Context("when setting vm metadata fails", func() {
			BeforeEach(func() {
				fakeCloud.SetVMMetadataError = errors.New("fake-set-metadata-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-set-metadata-error"))
			})

			It("still updates the current vm record", func() {
				_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
				Expect(err).To(HaveOccurred())
				Expect(fakeVMRepo.UpdateCurrentCID).To(Equal("fake-vm-cid"))
			})

			It("ignores not implemented error", func() {
				notImplementedCloudError := bicloud.NewCPIError("set_vm_metadata", bicloud.CmdError{
					Type:      "Bosh::Clouds::NotImplemented",
					Message:   "set_vm_metadata is not implemented by VCloudCloud::Cloud",
					OkToRetry: false,
				})
				fakeCloud.SetVMMetadataError = notImplementedCloudError

				_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when creating the vm fails", func() {
			BeforeEach(func() {
				fakeCloud.CreateVMErr = errors.New("fake-create-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(stemcell, deploymentManifest, diskCIDs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-error"))
			})
		})
	})
})
