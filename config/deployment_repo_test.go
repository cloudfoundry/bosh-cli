package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	. "github.com/cloudfoundry/bosh-init/config"
)

var _ = Describe("DeploymentRepo", func() {
	var (
		repo              DeploymentRepo
		configService     DeploymentConfigService
		fs                *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
		configService = NewFileSystemDeploymentConfigService(fs, fakeUUIDGenerator, logger)
		configService.SetConfigPath("/fake/path")
		repo = NewDeploymentRepo(configService)
	})

	Describe("UpdateCurrent", func() {
		It("updates deployment manifest sha1", func() {
			err := repo.UpdateCurrent("fake-manifest-sha1")
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := DeploymentFile{
				DirectorID:          "fake-uuid-0",
				CurrentManifestSHA1: "fake-manifest-sha1",
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})
	})

	Describe("FindCurrent", func() {
		Context("when a current manifest sha1 is set", func() {
			BeforeEach(func() {
				err := repo.UpdateCurrent("fake-manifest-sha1")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns current manifest sha1", func() {
				record, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(record).To(Equal("fake-manifest-sha1"))
			})
		})

		Context("when a current manifest sha1 is not set", func() {
			It("returns false", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
