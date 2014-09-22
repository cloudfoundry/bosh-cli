package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/stemcell"
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
			cid := CID("fake-cid")
			stemcell := Stemcell{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
			}
			err := repo.Save(stemcell, cid)
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := bmconfig.DeploymentConfig{
				Stemcells: []bmconfig.StemcellRecord{
					{
						Name:    "fake-name",
						Version: "fake-version",
						SHA1:    "fake-sha1",
						CID:     cid.String(),
					},
				},
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})
	})

	Describe("Find", func() {
		It("finds existing stemcell records", func() {
			expectedCid := CID("fake-cid")
			stemcell := Stemcell{
				Name:    "fake-name",
				Version: "fake-version",
				SHA1:    "fake-sha1",
			}
			repo.Save(stemcell, expectedCid)

			cid, found, err := repo.Find(stemcell)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(cid).To(Equal(expectedCid))
		})
	})
})
