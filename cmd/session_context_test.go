package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	fakeconf "github.com/cloudfoundry/bosh-init/cmd/config/fakes"
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
			ResolveTargetStub: func(in string) string { return in },
		}
		fs = fakesys.NewFakeFileSystem()
		context = nil
	})

	build := func() *SessionContextImpl { return NewSessionContextImpl(opts, config, fs) }

	Describe("Target", func() {
		It("returns resolved global option if provided", func() {
			config.TargetReturns("config-url")

			config.ResolveTargetStub = func(in string) string {
				Expect(in).To(Equal("opt-alias"))
				return "resolved-url"
			}

			opts.TargetOpt = "opt-alias"

			Expect(build().Target()).To(Equal("resolved-url"))
		})

		It("uses config value if no global option is provided", func() {
			config.TargetReturns("config-url")

			Expect(build().Target()).To(Equal("config-url"))
		})

		It("returns empty string if neither global option or config value is set", func() {
			Expect(build().Target()).To(Equal(""))
		})
	})

	Describe("Credentials", func() {
		It("defaults to config credentials for config target", func() {
			config.TargetReturns("config-url")

			config.CredentialsStub = func(target string) cmdconf.Creds {
				Expect(target).To(Equal("config-url"))
				return cmdconf.Creds{Username: "config-username"}
			}

			Expect(build().Credentials()).To(Equal(cmdconf.Creds{Username: "config-username"}))
		})

		It("prefers to use target from global option and returns config credentials", func() {
			config.CredentialsStub = func(target string) cmdconf.Creds {
				Expect(target).To(Equal("opt-url"))
				return cmdconf.Creds{Username: "config-username"}
			}

			opts.TargetOpt = "opt-url"

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
			opts.TargetOpt = "opt-url"
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

		It("uses config value for current target if no global option is provided", func() {
			config.CACertStub = func(target string) string {
				Expect(target).To(Equal("opt-url"))
				return "config-cert"
			}

			Expect(build().CACert()).To(Equal("config-cert"))
		})

		It("returns empty string if neither global option or config value is set", func() {
			Expect(build().CACert()).To(Equal(""))
		})
	})

	Describe("Deployment", func() {
		BeforeEach(func() {
			opts.TargetOpt = "opt-url"
		})

		It("returns global option if provided", func() {
			config.DeploymentReturns("config-dep")

			opts.DeploymentOpt = "opt-dep"

			Expect(build().Deployment()).To(Equal("opt-dep"))
		})

		It("uses config value for current target if no global option is provided", func() {
			config.DeploymentStub = func(target string) string {
				Expect(target).To(Equal("opt-url"))
				return "config-dep"
			}

			Expect(build().Deployment()).To(Equal("config-dep"))
		})

		It("returns empty string if neither global option or config value is set", func() {
			Expect(build().Deployment()).To(Equal(""))
		})
	})
})
