package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("DeploymentRepo", func() {
	var (
		repo          DeploymentRepo
		configService DeploymentConfigService
		fs            *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		configService = NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		repo = NewDeploymentRepo(configService)
	})

	Describe("UpdateCurrent", func() {
		It("updates deployment manifest sha1", func() {
			err := repo.UpdateCurrent("fake-manifest-sha1")
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := DeploymentConfig{
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
			It("returns current manifest sha1", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
