package config_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd/config"
)

var _ = Describe("NewFSConfigFromPath", func() {
	It("expands config path", func() {
		fs := fakesys.NewFakeFileSystem()
		fs.ExpandPathExpanded = "/expanded_config"

		config, err := NewFSConfigFromPath("/config", fs)
		Expect(err).ToNot(HaveOccurred())

		err = config.Save()
		Expect(err).ToNot(HaveOccurred())
		Expect(fs.FileExists("/expanded_config")).To(BeTrue())
	})

	It("returns empty config if file does not exist", func() {
		fs := fakesys.NewFakeFileSystem()

		config, err := NewFSConfigFromPath("/no_config", fs)
		Expect(err).ToNot(HaveOccurred())
		Expect(config.Target()).To(BeEmpty())
		Expect(config.Targets()).To(BeEmpty())
	})

	It("returns error if expanding path fails", func() {
		fs := fakesys.NewFakeFileSystem()
		fs.ExpandPathErr = errors.New("fake-err")

		_, err := NewFSConfigFromPath("/config", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})

	It("returns error if reading file fails", func() {
		fs := fakesys.NewFakeFileSystem()
		fs.WriteFileString("/config", "")
		fs.ReadFileError = errors.New("fake-err")

		_, err := NewFSConfigFromPath("/config", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})

	It("returns error if config file cannot be unmarshaled", func() {
		fs := fakesys.NewFakeFileSystem()
		fs.WriteFileString("/config", "-")

		_, err := NewFSConfigFromPath("/config", fs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("line 1"))
	})
})

var _ = Describe("FSConfig", func() {
	var (
		fs     *fakesys.FakeFileSystem
		config FSConfig
	)

	readConfig := func() FSConfig {
		config, err := NewFSConfigFromPath("/dir/sub-dir/config", fs)
		Expect(err).ToNot(HaveOccurred())

		return config
	}

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		config = readConfig()
	})

	Describe("Tagrets", func() {
		It("returns empty list if there are no remembered targets", func() {
			Expect(config.Targets()).To(BeEmpty())
		})

		It("returns list of previously remembered targets", func() {
			updatedConfig := config.SetTarget("url1", "", "")
			updatedConfig = updatedConfig.SetTarget("url2", "", "")
			updatedConfig = updatedConfig.SetTarget("url3", "alias3", "")
			Expect(updatedConfig.Targets()).To(Equal([]Target{
				Target{URL: "url1", Alias: ""},
				Target{URL: "url2", Alias: ""},
				Target{URL: "url3", Alias: "alias3"},
			}))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Targets()).To(Equal([]Target{
				Target{URL: "url1", Alias: ""},
				Target{URL: "url2", Alias: ""},
				Target{URL: "url3", Alias: "alias3"},
			}))
		})
	})

	Describe("SetTarget/Target/CACert", func() {
		It("returns empty if file does not exist", func() {
			Expect(config.Target()).To(Equal(""))
		})

		It("returns saved url", func() {
			updatedConfig := config.SetTarget("url", "", "")
			Expect(updatedConfig.Target()).To(Equal("url"))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Target()).To(Equal("url"))
		})

		It("returns saved url based on the alias, resolving to previously saved url", func() {
			updatedConfig := config.SetTarget("url1", "alias1", "")
			updatedConfig = updatedConfig.SetTarget("url2", "alias2", "")
			updatedConfig = updatedConfig.SetTarget("alias1", "", "")
			Expect(updatedConfig.Target()).To(Equal("url1"))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Target()).To(Equal("url1"))
		})

		It("saves empty CA certificate", func() {
			updatedConfig := config.SetTarget("url", "", "")
			Expect(updatedConfig.Target()).To(Equal("url"))
			Expect(updatedConfig.CACert("url")).To(Equal(""))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.CACert("url")).To(Equal(""))
		})

		// Valid from pov of FSConfig
		validCACert := "BEGIN\nca-cert"

		It("saves non-empty CA certificate and then unsets it", func() {
			updatedConfig := config.SetTarget("url", "", validCACert)
			Expect(updatedConfig.Target()).To(Equal("url"))
			Expect(updatedConfig.CACert("url")).To(Equal(validCACert))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.CACert("url")).To(Equal(validCACert))

			updatedConfig = reloadedConfig.SetTarget("url", "", "")
			Expect(updatedConfig.Target()).To(Equal("url"))
			Expect(updatedConfig.CACert("url")).To(Equal(""))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.CACert("url")).To(Equal(""))
		})

		It("saves CA cert via file path and does not need file system later", func() {
			fs.WriteFileString("/ca-cert", validCACert)

			updatedConfig := config.SetTarget("url", "", "/ca-cert")
			Expect(updatedConfig.CACert("url")).To(Equal(validCACert))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			err = fs.RemoveAll("/ca-cert")
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.CACert("url")).To(Equal(validCACert))
		})

		It("returns CA cert for alias", func() {
			updatedConfig := config.SetTarget("url", "alias", validCACert)
			Expect(updatedConfig.Target()).To(Equal("url"))
			Expect(updatedConfig.CACert("alias")).To(Equal(validCACert))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.CACert("alias")).To(Equal(validCACert))

			updatedConfig = reloadedConfig.SetTarget("url", "alias", "")
			Expect(updatedConfig.Target()).To(Equal("url"))
			Expect(updatedConfig.CACert("alias")).To(Equal(""))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.CACert("alias")).To(Equal(""))
		})

		It("does not update existing config when target is set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			Expect(updatedConfig.Target()).To(Equal("url"))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.Target()).To(Equal(""))
		})
	})

	Describe("ResolveTarget", func() {
		It("returns url if it's a known url", func() {
			updatedConfig := config.SetTarget("url", "", "")
			Expect(updatedConfig.ResolveTarget("url")).To(Equal("url"))
		})

		It("returns aliased url", func() {
			updatedConfig := config.SetTarget("url", "alias", "")
			updatedConfig = updatedConfig.SetTarget("url2", "alias2", "")
			Expect(updatedConfig.ResolveTarget("alias")).To(Equal("url"))
		})

		It("returns input when it's not an alias or url", func() {
			Expect(config.ResolveTarget("unknown")).To(Equal("unknown"))
		})
	})

	Describe("SetCredentials/Credentials/UnsetCredentials", func() {
		It("returns empty if target is not found", func() {
			Expect(config.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns empty if target is found but creds are not set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))

			updatedConfig = reloadedConfig.UnsetCredentials("url")
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns creds with username/password if target is found and basic creds are set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			updatedConfig = config.SetCredentials("url", Creds{Username: "user", Password: "pass"})
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{Username: "user", Password: "pass"}))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{Username: "user", Password: "pass"}))

			updatedConfig = reloadedConfig.UnsetCredentials("url")
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns creds with token if target is found and token is set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			updatedConfig = config.SetCredentials("url", Creds{RefreshToken: "token"})
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{RefreshToken: "token"}))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{RefreshToken: "token"}))

			updatedConfig = reloadedConfig.UnsetCredentials("url")
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns creds for alias if target is found and token is set", func() {
			updatedConfig := config.SetTarget("url", "alias", "")
			updatedConfig = config.SetCredentials("alias", Creds{RefreshToken: "token"})
			Expect(updatedConfig.Credentials("alias")).To(Equal(Creds{RefreshToken: "token"}))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("alias")).To(Equal(Creds{RefreshToken: "token"}))

			updatedConfig = reloadedConfig.UnsetCredentials("alias")
			Expect(updatedConfig.Credentials("alias")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Credentials("alias")).To(Equal(Creds{}))
		})

		It("does not update existing config when creds are set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			updatedConfig = config.SetCredentials("url", Creds{Username: "user"})
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{Username: "user"}))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.Credentials("url")).To(Equal(Creds{}))
		})
	})

	Describe("SetDeployment/Deployment", func() {
		It("returns empty if target is not found", func() {
			Expect(config.Deployment("url")).To(Equal(""))
		})

		It("returns empty if target is found but deployment is not set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			Expect(updatedConfig.Deployment("url")).To(Equal(""))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Deployment("url")).To(Equal(""))
		})

		It("returns deployment if target is found and deployment is set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			updatedConfig = config.SetDeployment("url", "deployment")
			Expect(updatedConfig.Deployment("url")).To(Equal("deployment"))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Deployment("url")).To(Equal("deployment"))
		})

		It("returns deployment for alias if target is found and deployment is set", func() {
			updatedConfig := config.SetTarget("url", "alias", "")
			updatedConfig = config.SetDeployment("alias", "deployment")
			Expect(updatedConfig.Deployment("alias")).To(Equal("deployment"))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(reloadedConfig.Deployment("alias")).To(Equal("deployment"))
		})

		It("does not update existing config when deployment is set", func() {
			updatedConfig := config.SetTarget("url", "", "")
			updatedConfig = config.SetDeployment("url", "deployment")
			Expect(updatedConfig.Deployment("url")).To(Equal("deployment"))

			err := updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.Deployment("url")).To(Equal(""))
		})
	})

	Describe("Save", func() {
		It("returns error if writing file fails", func() {
			fs.WriteFileError = errors.New("fake-err")

			config := readConfig()
			err := config.Save()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
