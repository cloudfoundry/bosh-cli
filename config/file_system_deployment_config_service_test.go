package config_test

import (
	"encoding/json"
	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("fileSystemConfigService", func() {
	var (
		service            DeploymentConfigService
		deploymentFilePath string
		fakeFs             *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		deploymentFilePath = "/some/deployment.json"
		logger := boshlog.NewLogger(boshlog.LevelNone)
		service = NewFileSystemDeploymentConfigService(deploymentFilePath, fakeFs, logger)
	})

	Describe("Load", func() {
		It("reads the given config file", func() {
			stemcells := []StemcellRecord{
				StemcellRecord{
					Name:    "fake-stemcell-name-1",
					Version: "fake-stemcell-version-1",
					SHA1:    "fake-stemcell-sha1-1",
					CID:     "fake-stemcell-cid-1",
				},
				StemcellRecord{
					Name:    "fake-stemcell-name-2",
					Version: "fake-stemcell-version-2",
					SHA1:    "fake-stemcell-sha1-2",
					CID:     "fake-stemcell-cid-2",
				},
			}
			deploymentFileContents, err := json.Marshal(map[string]interface{}{
				"uuid":      "deadbeef",
				"stemcells": stemcells,
				"vm_cid":    "fake-vm-cid",
				"disk_cid":  "fake-disk-cid",
			})
			fakeFs.WriteFile(deploymentFilePath, deploymentFileContents)

			config, err := service.Load()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.DeploymentUUID).To(Equal("deadbeef"))
			Expect(config.Stemcells).To(Equal(stemcells))
			Expect(config.VMCID).To(Equal("fake-vm-cid"))
			Expect(config.DiskCID).To(Equal("fake-disk-cid"))
		})

		Context("when the config does not exist", func() {
			It("returns an empty DeploymentConfig", func() {
				config, err := service.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(Equal(DeploymentConfig{}))
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
				Expect(err.Error()).To(ContainSubstring("Unmarshalling deployment config file `/some/deployment.json'"))
				Expect(config).To(Equal(DeploymentConfig{}))
			})
		})
	})

	Describe("Save", func() {
		It("writes the deployment config to the deployment file", func() {
			config := DeploymentConfig{
				DeploymentUUID: "deadbeef",
				Stemcells: []StemcellRecord{
					{
						Name:    "fake-stemcell-name",
						Version: "fake-stemcell-version",
						SHA1:    "fake-stemcell-sha1",
						CID:     "fake-stemcell-cid",
					},
				},
				VMCID:   "fake-vm-cid",
				DiskCID: "fake-disk-cid",
			}

			err := service.Save(config)
			Expect(err).NotTo(HaveOccurred())

			deploymentFileContents, err := fakeFs.ReadFileString(deploymentFilePath)
			deploymentFile := DeploymentFile{
				UUID: "deadbeef",
				Stemcells: []StemcellRecord{
					{
						Name:    "fake-stemcell-name",
						Version: "fake-stemcell-version",
						SHA1:    "fake-stemcell-sha1",
						CID:     "fake-stemcell-cid",
					},
				},
				VMCID:   "fake-vm-cid",
				DiskCID: "fake-disk-cid",
			}
			expectedDeploymentFileContents, err := json.MarshalIndent(deploymentFile, "", "    ")
			Expect(deploymentFileContents).To(Equal(string(expectedDeploymentFileContents)))
		})

		Context("when the deployment file cannot be written", func() {
			BeforeEach(func() {
				fakeFs.WriteToFileError = errors.New("")
			})

			It("returns an error when it cannot write the config file", func() {
				config := DeploymentConfig{
					Stemcells: []StemcellRecord{},
				}
				err := service.Save(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Writing deployment config file `/some/deployment.json'"))
			})
		})
	})
})
