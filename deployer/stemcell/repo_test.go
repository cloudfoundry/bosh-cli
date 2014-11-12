package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
)

var _ = Describe("Repo", func() {
	var (
		repo          Repo
		configService bmconfig.DeploymentConfigService
		fs            *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		configService = bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		repo = NewRepo(configService)
	})

	Describe("Save", func() {
		It("saves the stemcell record using the config service", func() {
			stemcell := CloudStemcell{CID: "fake-cid"}
			stemcellManifest := Manifest{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
			}
			err := repo.Save(stemcellManifest, stemcell)
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := bmconfig.DeploymentConfig{
				Stemcells: []bmconfig.StemcellRecord{
					{
						Name:    "fake-name",
						Version: "fake-version",
						SHA1:    "fake-sha1",
						CID:     "fake-cid",
					},
				},
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})
	})

	Describe("Find", func() {
		It("finds existing stemcell records", func() {
			expectedStemcell := CloudStemcell{CID: "fake-cid"}
			stemcellManifest := Manifest{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
			}
			repo.Save(stemcellManifest, expectedStemcell)

			stemcell, found, err := repo.Find(stemcellManifest)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(stemcell).To(Equal(expectedStemcell))
		})
	})
})
