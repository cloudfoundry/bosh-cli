package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	fakeconf "github.com/cloudfoundry/bosh-cli/cmd/config/configfakes"
)

var _ = Describe("SessionContextImpl", func() {
	var (
		opts    BoshOpts
		config  *fakeconf.FakeConfig
		fs      *fakesys.FakeFileSystem
		context *SessionContextImpl
	)

	BeforeEach(func() {
		opts = BoshOpts{}
		config = &fakeconf.FakeConfig{
			ResolveEnvironmentStub: func(in string) string { return in },
		}
		fs = fakesys.NewFakeFileSystem()
		context = nil
	})

	build := func() *SessionContextImpl { return NewSessionContextImpl(opts, config, fs) }

	Describe("Environment", func() {
		It("returns resolved global option if provided", func() {
			config.ResolveEnvironmentStub = func(in string) string {
				Expect(in).To(Equal("opt-alias"))
				return "resolved-url"
			}

			opts.EnvironmentOpt = "opt-alias"

			Expect(build().Environment()).To(Equal("resolved-url"))
		})

		It("returns empty string if global option is not set", func() {
			Expect(build().Environment()).To(Equal(""))
		})
	})

	Describe("Credentials", func() {
		It("defaults to config credentials for environment global option", func() {
			config.CredentialsStub = func(environment string) cmdconf.Creds {
				Expect(environment).To(Equal("opt-alias"))
				return cmdconf.Creds{Username: "config-username"}
			}

			opts.EnvironmentOpt = "opt-alias"

			Expect(build().Credentials()).To(Equal(cmdconf.Creds{Username: "config-username"}))
		})

		It("overrides username with username global option", func() {
			config.CredentialsReturns(cmdconf.Creds{Username: "config-username"})

			opts.UsernameOpt = "opt-username"

			Expect(build().Credentials()).To(Equal(cmdconf.Creds{Username: "opt-username"}))
		})

		It("overrides password with password global option", func() {
			config.CredentialsReturns(cmdconf.Creds{Password: "config-password"})

			opts.PasswordOpt = "opt-password"

			Expect(build().Credentials()).To(Equal(cmdconf.Creds{Password: "opt-password"}))
		})

		It("overrides uaa client and resets secret if uaa client global option is provided", func() {
			config.CredentialsReturns(cmdconf.Creds{
				Client:       "config-client",
				ClientSecret: "config-client-secret",
			})

			opts.UAAClientOpt = "opt-client"

			Expect(build().Credentials()).To(Equal(cmdconf.Creds{
				Client:       "opt-client",
				ClientSecret: "",
			}))
		})

		It("overrides uaa client and secret if uaa client global option is provided", func() {
			config.CredentialsReturns(cmdconf.Creds{
				Client:       "config-client",
				ClientSecret: "config-client-secret",
			})

			opts.UAAClientOpt = "opt-client"
			opts.UAAClientSecretOpt = "opt-client-secret"

			Expect(build().Credentials()).To(Equal(cmdconf.Creds{
				Client:       "opt-client",
				ClientSecret: "opt-client-secret",
			}))
		})
	})

	Describe("CACert", func() {
		BeforeEach(func() {
			opts.EnvironmentOpt = "opt-url"
		})

		It("returns global option if provided as non-file-path", func() {
			config.CACertReturns("config-cert")

			opts.CACertOpt = "opt-cert"

			Expect(build().CACert()).To(Equal("opt-cert"))
		})

		It("returns global option as value if provided as file path", func() {
			fs.WriteFileString("/cert", "file-cert")

			config.CACertReturns("config-cert")

			opts.CACertOpt = "/cert"

			Expect(build().CACert()).To(Equal("file-cert"))
		})

		It("returns empty value if provided file path cannot be read", func() {
			fs.WriteFileString("/cert", "file-cert")
			fs.ReadFileError = errors.New("fake-err")

			config.CACertReturns("config-cert")

			opts.CACertOpt = "/cert"

			Expect(build().CACert()).To(Equal(""))
		})

		It("returns empty string if global option is not set", func() {
			Expect(build().CACert()).To(Equal(""))
		})
	})

	Describe("Deployment", func() {
		It("returns global option if provided", func() {
			opts.DeploymentOpt = "opt-dep"
			Expect(build().Deployment()).To(Equal("opt-dep"))
		})

		It("returns empty string if global option is not set", func() {
			Expect(build().Deployment()).To(Equal(""))
		})
	})
})
