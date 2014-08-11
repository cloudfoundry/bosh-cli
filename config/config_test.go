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
		service        Service
		configFilePath string
		fakeFs         *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		configFilePath = "/somepath"
		logger := boshlog.NewLogger(boshlog.LevelNone)
		service = NewFileSystemConfigService(logger, fakeFs, configFilePath)
	})

	Describe("Load", func() {
		It("reads the given config file", func() {
			configFileContents, err := json.MarshalIndent(Config{Deployment: "/some/path"}, "", "    ")
			fakeFs.WriteFile(configFilePath, configFileContents)

			config, err := service.Load()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Deployment).To(Equal("/some/path"))
		})

		Context("when the config file does not exist", func() {
			It("returns an empty Config", func() {
				config, err := service.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(config).To(Equal(Config{}))
			})
		})

		Context("when the config file is invalid", func() {
			It("returns an empty Config and an error", func() {
				fakeFs.WriteFileString(configFilePath, "invalid json")
				config, err := service.Load()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unmarshalling JSON config file `/somepath'"))
				Expect(config).To(Equal(Config{}))
			})
		})
	})

	Describe("Save", func() {
		It("writes the given config to the config file", func() {
			config := Config{Deployment: "/some/path"}

			err := service.Save(config)
			Expect(err).NotTo(HaveOccurred())

			configFileContents, err := fakeFs.ReadFileString(configFilePath)
			expectedConfigFileContents, err := json.MarshalIndent(config, "", "    ")
			Expect(configFileContents).To(Equal(string(expectedConfigFileContents)))
		})

		Context("when the config file cannot be written", func() {
			BeforeEach(func() {
				fakeFs.WriteToFileError = errors.New("")
			})

			It("returns an error when it cannot write the config file", func() {
				config := Config{Deployment: "/some/path"}
				err := service.Save(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Writing config file `/somepath'"))
			})
		})
	})
})
