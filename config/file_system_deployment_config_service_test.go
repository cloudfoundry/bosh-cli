package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/config"

	"encoding/json"
	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
)

var _ = Describe("fileSystemConfigService", func() {
	var (
		service            DeploymentConfigService
		deploymentFilePath string
		fakeFs             *fakesys.FakeFileSystem
		fakeUUIDGenerator  *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		deploymentFilePath = "/some/deployment.json"
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
		service = NewFileSystemDeploymentConfigService(deploymentFilePath, fakeFs, fakeUUIDGenerator, logger)
	})

	Describe("Exists", func() {
		It("returns true if the config file exists", func() {
			fakeFs.WriteFileString(deploymentFilePath, "")
			Expect(service.Exists()).To(BeTrue())
		})

		It("returns false if the config file does not exist", func() {
			Expect(service.Exists()).To(BeFalse())
		})
	})

	Describe("Load", func() {
		It("reads the given config file", func() {
			stemcells := []StemcellRecord{
				StemcellRecord{
					Name:    "fake-stemcell-name-1",
					Version: "fake-stemcell-version-1",
					CID:     "fake-stemcell-cid-1",
				},
				StemcellRecord{
					Name:    "fake-stemcell-name-2",
					Version: "fake-stemcell-version-2",
					CID:     "fake-stemcell-cid-2",
				},
			}
			disks := []DiskRecord{
				{
					ID:   "fake-disk-id",
					CID:  "fake-disk-cid",
					Size: 1024,
					CloudProperties: bmproperty.Map{
						"fake-disk-property-key": "fake-disk-property-value",
					},
				},
			}
			deploymentFileContents, err := json.Marshal(bmproperty.Map{
				"director_id":     "fake-director-id",
				"deployment_id":   "fake-deployment-id",
				"stemcells":       stemcells,
				"current_vm_cid":  "fake-vm-cid",
				"current_disk_id": "fake-disk-id",
				"disks":           disks,
			})
			fakeFs.WriteFile(deploymentFilePath, deploymentFileContents)

			deploymentFile, err := service.Load()
			Expect(err).NotTo(HaveOccurred())
			Expect(deploymentFile.DirectorID).To(Equal("fake-director-id"))
			Expect(deploymentFile.Stemcells).To(Equal(stemcells))
			Expect(deploymentFile.CurrentVMCID).To(Equal("fake-vm-cid"))
			Expect(deploymentFile.CurrentDiskID).To(Equal("fake-disk-id"))
			Expect(deploymentFile.Disks).To(Equal(disks))
		})

		Context("when the config does not exist", func() {
			It("returns a new DeploymentConfig with generated defaults", func() {
				config, err := service.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(Equal(DeploymentFile{
					DirectorID: "fake-uuid-0",
				}))

				Expect(fakeFs.FileExists(deploymentFilePath)).To(BeTrue())
			})
		})

		Context("when reading config file fails", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(deploymentFilePath, "{}")
				fakeFs.ReadFileError = errors.New("fake-read-error")
			})

			It("returns an error", func() {
				_, err := service.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-read-error"))
			})
		})

		Context("when the config is invalid", func() {
			It("returns an empty DeploymentConfig and an error", func() {
				fakeFs.WriteFileString(deploymentFilePath, "some invalid content")
				config, err := service.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshalling deployment config file '/some/deployment.json'"))
				Expect(config).To(Equal(DeploymentFile{}))
			})
		})
	})

	Describe("Save", func() {
		It("writes the deployment config to the deployment file", func() {
			config := DeploymentFile{
				DirectorID: "deadbeef",
				Stemcells: []StemcellRecord{
					{
						Name:    "fake-stemcell-name",
						Version: "fake-stemcell-version",
						CID:     "fake-stemcell-cid",
					},
				},
				CurrentVMCID: "fake-vm-cid",
				Disks: []DiskRecord{
					{
						CID:  "fake-disk-cid",
						Size: 1024,
						CloudProperties: bmproperty.Map{
							"fake-disk-property-key": "fake-disk-property-value",
						},
					},
				},
			}

			err := service.Save(config)
			Expect(err).NotTo(HaveOccurred())

			deploymentFileContents, err := fakeFs.ReadFileString(deploymentFilePath)
			deploymentFile := DeploymentFile{
				DirectorID: "deadbeef",
				Stemcells: []StemcellRecord{
					{
						Name:    "fake-stemcell-name",
						Version: "fake-stemcell-version",
						CID:     "fake-stemcell-cid",
					},
				},
				CurrentVMCID: "fake-vm-cid",
				Disks: []DiskRecord{
					{
						CID:  "fake-disk-cid",
						Size: 1024,
						CloudProperties: bmproperty.Map{
							"fake-disk-property-key": "fake-disk-property-value",
						},
					},
				},
			}
			expectedDeploymentFileContents, err := json.MarshalIndent(deploymentFile, "", "    ")
			Expect(deploymentFileContents).To(Equal(string(expectedDeploymentFileContents)))
		})

		Context("when the deployment file cannot be written", func() {
			BeforeEach(func() {
				fakeFs.WriteToFileError = errors.New("")
			})

			It("returns an error when it cannot write the config file", func() {
				config := DeploymentFile{
					Stemcells: []StemcellRecord{},
				}
				err := service.Save(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Writing deployment config file '/some/deployment.json'"))
			})
		})
	})
})
