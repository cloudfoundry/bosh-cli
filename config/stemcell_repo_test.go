package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("Repo", func() {
	var (
		repo          StemcellRepo
		configService DeploymentConfigService
		fs            *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		configService = NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		repo = NewStemcellRepo(configService)
	})

	Describe("Save", func() {
		It("saves the stemcell record using the config service", func() {
			stemcellRecord := StemcellRecord{
				Name:    "fake-name",
				Version: "fake-version",
				CID:     "fake-cid",
			}
			err := repo.Save(stemcellRecord)
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := DeploymentConfig{
				Stemcells: []StemcellRecord{
					{
						Name:    "fake-name",
						Version: "fake-version",
						CID:     "fake-cid",
					},
				},
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})
	})

	Describe("Find", func() {
		It("finds existing stemcell records", func() {
			stemcellRecord := StemcellRecord{
				Name:    "fake-name",
				Version: "fake-version",
				CID:     "fake-cid",
			}
			repo.Save(stemcellRecord)

			foundStemcellRecord, found, err := repo.Find("fake-name", "fake-version")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(foundStemcellRecord).To(Equal(stemcellRecord))
		})
	})
})
