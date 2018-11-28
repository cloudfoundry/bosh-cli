package config_test

import (
	"errors"
	"os"

	. "github.com/cloudfoundry/bosh-cli/cmd/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/uaa"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
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
		Expect(config.Environments()).To(BeEmpty())
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
		config Config
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

	Describe("Environments", func() {
		It("returns empty list if there are no remembered environments", func() {
			Expect(config.Environments()).To(BeEmpty())
		})

		It("returns list of previously remembered environments", func() {
			updatedConfig, err := config.AliasEnvironment("url1", "alias1", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig, err = updatedConfig.AliasEnvironment("url2", "alias2", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig, err = updatedConfig.AliasEnvironment("url3", "alias3", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.Environments()).To(Equal([]Environment{
				Environment{URL: "url1", Alias: "alias1"},
				Environment{URL: "url2", Alias: "alias2"},
				Environment{URL: "url3", Alias: "alias3"},
			}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.Environments()).To(Equal([]Environment{
				Environment{URL: "url1", Alias: "alias1"},
				Environment{URL: "url2", Alias: "alias2"},
				Environment{URL: "url3", Alias: "alias3"},
			}))
		})
	})

	Describe("DeleteAlias", func() {
		var findEnv = func(alias string, envs []Environment) *Environment {
			for i := range envs {
				if envs[i].Alias == alias {
					return &envs[i]
				}
			}
			return nil
		}

		BeforeEach(func() {
			var err error
			config, err = config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			config, err = config.AliasEnvironment("url2", "alias2", "")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if the alias does not exist", func() {
			_, err := config.UnaliasEnvironment("alias-does-not-exist")
			Expect(err).To(HaveOccurred())
		})

		It("deletes an environment with a particular alias", func() {
			len := len(config.Environments())
			var err error
			config, err = config.UnaliasEnvironment("alias")
			Expect(err).ToNot(HaveOccurred())

			envs := config.Environments()
			Expect(envs).To(HaveLen(len - 1))

			Expect(findEnv("alias", envs)).To(BeNil())
			Expect(findEnv("alias2", envs)).ToNot(BeNil())
		})
	})

	Describe("AliasEnvironment/CACert", func() {
		It("returns empty if file does not exist", func() {
			Expect(config.CACert("url")).To(Equal(""))
		})

		It("returns error if url is empty", func() {
			_, err := config.AliasEnvironment("", "alias", "ca-cert")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty environment URL"))
		})

		It("overwrites when an entry with the given url is already present", func() {
			config, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			config, err = config.AliasEnvironment("url", "different-alias", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(config.Environments()).To(HaveLen(1))
			Expect(config.Environments()[0].Alias).To(Equal("different-alias"))
			Expect(config.Environments()[0].URL).To(Equal("url"))

		})

		It("overwrites whent an entry with the given alias is already present", func() {
			config, err := config.AliasEnvironment("url", "alias", "ca")
			Expect(err).ToNot(HaveOccurred())

			config, err = config.AliasEnvironment("different-url", "alias", "diff-ca")
			Expect(err).ToNot(HaveOccurred())

			Expect(config.Environments()).To(HaveLen(1))
			Expect(config.Environments()[0].Alias).To(Equal("alias"))
			Expect(config.Environments()[0].URL).To(Equal("different-url"))
		})

		It("returns error if alias is empty", func() {
			_, err := config.AliasEnvironment("url", "", "ca-cert")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty environment alias"))
		})

		It("returns saved url", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("url")).To(Equal("url"))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.ResolveEnvironment("url")).To(Equal("url"))
		})

		It("returns saved url based on the alias", func() {
			updatedConfig, err := config.AliasEnvironment("url1", "alias1", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig, err = updatedConfig.AliasEnvironment("url2", "alias2", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("alias1")).To(Equal("url1"))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()

			Expect(reloadedConfig.ResolveEnvironment("alias1")).To(Equal("url1"))
		})

		It("saves empty CA certificate", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("url")).To(Equal("url"))
			Expect(updatedConfig.CACert("url")).To(Equal(""))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()

			Expect(reloadedConfig.CACert("url")).To(Equal(""))
		})

		It("saves non-empty CA certificate and then unsets it", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "ca-cert")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("url")).To(Equal("url"))
			Expect(updatedConfig.CACert("url")).To(Equal("ca-cert"))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.CACert("url")).To(Equal("ca-cert"))

			updatedConfig, err = reloadedConfig.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("url")).To(Equal("url"))
			Expect(updatedConfig.CACert("url")).To(Equal(""))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(reloadedConfig.CACert("url")).To(Equal(""))
		})

		It("returns CA cert for alias", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "ca-cert")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("url")).To(Equal("url"))
			Expect(updatedConfig.CACert("alias")).To(Equal("ca-cert"))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.CACert("alias")).To(Equal("ca-cert"))

			updatedConfig, err = reloadedConfig.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("url")).To(Equal("url"))
			Expect(updatedConfig.CACert("alias")).To(Equal(""))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(reloadedConfig.CACert("alias")).To(Equal(""))
		})
	})

	Describe("ResolveEnvironment", func() {
		It("returns url if it's a known url", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("url")).To(Equal("url"))
		})

		It("returns aliased url", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig, err = updatedConfig.AliasEnvironment("url2", "alias2", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.ResolveEnvironment("alias")).To(Equal("url"))
		})

		It("returns input when it's not an alias or url", func() {
			Expect(config.ResolveEnvironment("unknown")).To(Equal("unknown"))
		})

		It("returns empty when alias or url is empty", func() {
			Expect(config.ResolveEnvironment("")).To(Equal(""))
		})
	})

	Describe("UpdateConfigWithToken", func() {
		var envAlias string

		BeforeEach(func() {
			envAlias = "test-env"
		})

		It("updates based on a non-refreshable token and saves", func() {
			err := config.UpdateConfigWithToken(envAlias, uaa.NewAccessToken("next-access-type", "next-access-token"))
			Expect(err).ToNot(HaveOccurred())
			reloadedConfig := readConfig()
			Expect(reloadedConfig.Credentials(envAlias)).To(Equal(Creds{
				AccessToken:     "next-access-token",
				AccessTokenType: "next-access-type",
				RefreshToken:    "",
			}))
		})

		It("updates based on a refreshable token and saves", func() {
			err := config.UpdateConfigWithToken(envAlias, uaa.NewRefreshableAccessToken("next-access-type", "next-access-token", "next-refresh-token"))
			Expect(err).ToNot(HaveOccurred())
			reloadedConfig := readConfig()
			Expect(reloadedConfig.Credentials(envAlias)).To(Equal(Creds{
				AccessToken:     "next-access-token",
				AccessTokenType: "next-access-type",
				RefreshToken:    "next-refresh-token",
			}))
		})

		It("returns an error when save fails", func() {
			fs.WriteFileError = errors.New("write error")

			err := config.UpdateConfigWithToken(envAlias, uaa.NewRefreshableAccessToken("next-access-type", "next-access-token", "next-refresh-token"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("write error"))
		})
	})

	Describe("SetCredentials/Credentials/UnsetCredentials", func() {
		It("returns empty if environment is not found", func() {
			Expect(config.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns empty if environment is found but creds are not set", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))

			updatedConfig = reloadedConfig.UnsetCredentials("url")
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns creds with username/password if environment is found and basic creds are set", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig = config.SetCredentials("url", Creds{Client: "user", ClientSecret: "pass"})
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{Client: "user", ClientSecret: "pass"}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{Client: "user", ClientSecret: "pass"}))

			updatedConfig = reloadedConfig.UnsetCredentials("url")
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns creds with token if environment is found and token is set", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig = config.SetCredentials("url", Creds{AccessToken: "access", AccessTokenType: "access-type", RefreshToken: "token"})
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{AccessToken: "access", AccessTokenType: "access-type", RefreshToken: "token"}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{AccessToken: "access", AccessTokenType: "access-type", RefreshToken: "token"}))

			updatedConfig = reloadedConfig.UnsetCredentials("url")
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(reloadedConfig.Credentials("url")).To(Equal(Creds{}))
		})

		It("returns creds for alias if environment is found and token is set", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig = config.SetCredentials("alias", Creds{AccessToken: "access", AccessTokenType: "access-type", RefreshToken: "token"})
			Expect(updatedConfig.Credentials("alias")).To(Equal(Creds{AccessToken: "access", AccessTokenType: "access-type", RefreshToken: "token"}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig := readConfig()
			Expect(reloadedConfig.Credentials("alias")).To(Equal(Creds{AccessToken: "access", AccessTokenType: "access-type", RefreshToken: "token"}))

			updatedConfig = reloadedConfig.UnsetCredentials("alias")
			Expect(updatedConfig.Credentials("alias")).To(Equal(Creds{}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			reloadedConfig = readConfig()
			Expect(reloadedConfig.Credentials("alias")).To(Equal(Creds{}))
		})

		It("does not update existing config when creds are set", func() {
			updatedConfig, err := config.AliasEnvironment("url", "alias", "")
			Expect(err).ToNot(HaveOccurred())

			updatedConfig = config.SetCredentials("url", Creds{Client: "user"})
			Expect(updatedConfig.Credentials("url")).To(Equal(Creds{Client: "user"}))

			err = updatedConfig.Save()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.Credentials("url")).To(Equal(Creds{}))
		})
	})

	Describe("Save", func() {
		It("chmods the file to 600", func() {
			config := readConfig()
			err := config.Save()
			Expect(err).ToNot(HaveOccurred())
			fileInfo, _ := fs.Stat("/dir/sub-dir/config")
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0600)))
		})

		It("returns error if chmoding file fails", func() {
			fs.ChmodErr = errors.New("chmod error")

			config := readConfig()
			err := config.Save()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("chmod error"))
		})

		It("returns error if writing file fails", func() {
			fs.WriteFileError = errors.New("write error")

			config := readConfig()
			err := config.Save()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("write error"))
		})
	})
})
