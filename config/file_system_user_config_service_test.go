package config_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("FileSystemUserConfigService", func() {
	var (
		configPath    string
		fs            *fakesys.FakeFileSystem
		configService UserConfigService
	)

	BeforeEach(func() {
		configPath = "/fake-path"
		fs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		configService = NewFileSystemUserConfigService(configPath, fs, logger)
	})

	It("loads the user config from file system", func() {
		configFileContents := `
{ 
	"deployment": "expected-deployment-file"
}
`
		err := fs.WriteFileString(configPath, configFileContents)
		Expect(err).ToNot(HaveOccurred())

		config, err := configService.Load()
		Expect(err).ToNot(HaveOccurred())
		Expect(config.DeploymentManifestPath).To(Equal("expected-deployment-file"))
	})

	It("saves the user config to the file system", func() {
		config := UserConfig{
			DeploymentManifestPath: "/fake-path",
		}

		err := configService.Save(config)
		Expect(err).NotTo(HaveOccurred())

		savedConfig, err := configService.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(savedConfig).To(Equal(config))
	})
})
