package cmd_test

import (
	"errors"
	"os"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

var _ = Describe("Factory", func() {
	var (
		fs           *fakesys.FakeFileSystem
		factory      Factory
		fakeFilePath string
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()

		ui := boshui.NewConfUI(logger)
		defer ui.Flush()

		deps := NewBasicDeps(ui, logger)
		deps.FS = fs

		factory = NewFactory(deps)
		fakeFilePath = filepath.Join("/", "file")
	})

	Describe("unknown commands, args and flags", func() {
		BeforeEach(func() {
			err := fs.WriteFileString(filepath.Join("/", "file"), "")
			Expect(err).ToNot(HaveOccurred())
		})

		cmds := map[string][]string{
			"help":                  []string{},
			"add-blob":              []string{filepath.Join("/", "file"), "directory"},
			"attach-disk":           []string{"instance/abad1dea", "disk-cid-123"},
			"blobs":                 []string{},
			"interpolate":           []string{filepath.Join("/", "file")},
			"cancel-task":           []string{"1234"},
			"clean-up":              []string{},
			"cloud-check":           []string{},
			"cloud-config":          []string{},
			"create-env":            []string{filepath.Join("/", "file")},
			"sha2ify-release":       []string{filepath.Join("/", "file"), filepath.Join("/", "file2")},
			"create-release":        []string{filepath.Join("/", "file")},
			"delete-deployment":     []string{},
			"delete-disk":           []string{"cid"},
			"delete-env":            []string{filepath.Join("/", "file")},
			"delete-release":        []string{"release-version"},
			"delete-snapshot":       []string{"cid"},
			"delete-snapshots":      []string{},
			"delete-stemcell":       []string{"name/version"},
			"delete-vm":             []string{"cid"},
			"deploy":                []string{filepath.Join("/", "file")},
			"deployment":            []string{},
			"deployments":           []string{},
			"disks":                 []string{},
			"alias-env":             []string{"alias"},
			"environment":           []string{},
			"environments":          []string{},
			"errands":               []string{},
			"events":                []string{},
			"export-release":        []string{"release/version", "os/version"},
			"finalize-release":      []string{filepath.Join("/", "file")},
			"generate-job":          []string{filepath.Join("/", "file")},
			"generate-package":      []string{filepath.Join("/", "file")},
			"init-release":          []string{},
			"inspect-release":       []string{"name/version"},
			"instances":             []string{},
			"locks":                 []string{},
			"log-in":                []string{},
			"log-out":               []string{},
			"logs":                  []string{"slug"},
			"manifest":              []string{},
			"recreate":              []string{"slug"},
			"releases":              []string{},
			"remove-blob":           []string{filepath.Join("/", "file")},
			"reset-release":         []string{},
			"restart":               []string{"slug"},
			"run-errand":            []string{"name"},
			"runtime-config":        []string{},
			"snapshots":             []string{"group/id"},
			"start":                 []string{"slug"},
			"stemcells":             []string{},
			"stop":                  []string{"slug"},
			"sync-blobs":            []string{},
			"take-snapshot":         []string{"group/id"},
			"task":                  []string{"1234"},
			"tasks":                 []string{},
			"update-cloud-config":   []string{filepath.Join("/", "file")},
			"update-resurrection":   []string{"off"},
			"update-runtime-config": []string{filepath.Join("/", "file")},
			"upload-blobs":          []string{},
			"upload-release":        []string{filepath.Join("/", "file")},
			"upload-stemcell":       []string{filepath.Join("/", "file")},
			"vms":                   []string{},
		}

		for cmd, requiredArgs := range cmds {
			cmd, requiredArgs := cmd, requiredArgs // copy

			Describe(cmd, func() {
				It("fails with extra arguments", func() {
					cmdWithArgs := append([]string{cmd}, requiredArgs...)
					cmdWithArgs = append(cmdWithArgs, "extra", "args")

					_, err := factory.New(cmdWithArgs)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("does not support extra arguments: extra, args"))
				})
			})
		}

		Describe("ssh", func() {
			It("uses all remaining arguments as a command", func() {
				cmd, err := factory.New([]string{"ssh", "group", "cmd", "extra", "args"})
				Expect(err).ToNot(HaveOccurred())

				opts := cmd.Opts.(*SSHOpts)
				Expect(opts.Command).To(Equal([]string{"cmd", "extra", "args"}))
			})

			It("uses all remaining arguments as a command even that look like flags", func() {
				cmd, err := factory.New([]string{"ssh", "group", "cmd", "extra", "args", "--", "--gw-disable"})
				Expect(err).ToNot(HaveOccurred())

				opts := cmd.Opts.(*SSHOpts)
				Expect(opts.Command).To(Equal([]string{"cmd", "extra", "args", "--gw-disable"}))
			})

			It("returns error if command is given and extra arguments are specified", func() {
				_, err := factory.New([]string{"ssh", "group", "-c", "command", "--", "extra", "args"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("does not support extra arguments: extra, args"))
			})
		})

		It("catches unknown commands and lists available commands", func() {
			_, err := factory.New([]string{"unknown-cmd"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unknown command `unknown-cmd'. Please specify one command of: add-blob"))
		})

		It("catches unknown global flags", func() {
			_, err := factory.New([]string{"--unknown-flag"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unknown flag `unknown-flag'"))
		})

		It("catches unknown command flags", func() {
			_, err := factory.New([]string{"ssh", "--unknown-flag"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unknown flag `unknown-flag'"))
		})
	})

	Describe("gateway flags", func() {
		It("ssh command has configured gateway flags", func() {
			cmd, err := factory.New([]string{"ssh", "group", "cmd", "extra", "args", "--", "--gw-disable"})
			Expect(err).ToNot(HaveOccurred())

			_, _, err = cmd.Opts.(*SSHOpts).GatewayFlags.AsSSHOpts()
			Expect(err).ToNot(HaveOccurred())
		})

		It("scp command has configured gateway flags", func() {
			cmd, err := factory.New([]string{"scp", "group", "cmd", "extra", "args", "--", "--gw-disable"})
			Expect(err).ToNot(HaveOccurred())

			_, _, err = cmd.Opts.(*SCPOpts).GatewayFlags.AsSSHOpts()
			Expect(err).ToNot(HaveOccurred())
		})

		It("logs -f command has configured gateway flags", func() {
			cmd, err := factory.New([]string{"logs", "-f", "cmd"})
			Expect(err).ToNot(HaveOccurred())

			_, _, err = cmd.Opts.(*LogsOpts).GatewayFlags.AsSSHOpts()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("deploy command", func() {
		BeforeEach(func() {
			err := fs.WriteFileString(fakeFilePath, "")
			Expect(err).ToNot(HaveOccurred())
		})

		It("parses multiple skip-drain flags", func() {
			cmd, err := factory.New([]string{"deploy", "--skip-drain=job1", "--skip-drain=job2", fakeFilePath})
			Expect(err).ToNot(HaveOccurred())

			slug1, _ := boshdir.NewInstanceGroupOrInstanceSlugFromString("job1")
			slug2, _ := boshdir.NewInstanceGroupOrInstanceSlugFromString("job2")

			opts := cmd.Opts.(*DeployOpts)
			Expect(opts.SkipDrain).To(Equal([]boshdir.SkipDrain{
				{Slug: slug1},
				{Slug: slug2},
			}))
		})

		It("errors when excluding = from --skip-drain", func() {
			_, err := factory.New([]string{"deploy", "--skip-drain", "job1", fakeFilePath})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Not found: open job1: no such file or directory"))
		})

		It("defaults --skip-drain option value to all", func() {
			cmd, err := factory.New([]string{"deploy", "--skip-drain", fakeFilePath})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*DeployOpts)
			Expect(opts.SkipDrain).To(Equal([]boshdir.SkipDrain{
				{All: true},
			}))
		})
	})

	Describe("create-env command (command that uses FileBytesArg)", func() {
		It("returns *nice error from FileBytesArg* error if it cannot read manifest", func() {
			fs.ReadFileError = errors.New("fake-err")

			_, err := factory.New([]string{"create-env", "manifest.yml"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open manifest.yml: no such file or directory"))
		})

		It("is configured with config server flags for vars store", func() {
			err := fs.WriteFileString(fakeFilePath, "")
			Expect(err).ToNot(HaveOccurred())

			configServerTLSConfig := NewFakeTLSConfig()

			cmd, err := factory.New([]string{
				"create-env", fakeFilePath,
				"--config-server-url", "config-server-url",
				"--config-server-tls-ca", configServerTLSConfig.CA,
				"--config-server-tls-certificate", configServerTLSConfig.Certificate,
				"--config-server-tls-private-key", configServerTLSConfig.PrivateKey,
				"--config-server-namespace", "config-server-ns",
				"--vars-store", "config-server://",
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = cmd.Opts.(*CreateEnvOpts).VarsStore.List()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Listing of variables in config server is not supported"))
		})
	})

	Describe("delete-env command", func() {
		It("is configured with config server flags for vars store", func() {
			err := fs.WriteFileString(fakeFilePath, "")
			Expect(err).ToNot(HaveOccurred())

			configServerTLSConfig := NewFakeTLSConfig()

			cmd, err := factory.New([]string{
				"delete-env", fakeFilePath,
				"--config-server-url", "config-server-url",
				"--config-server-tls-ca", configServerTLSConfig.CA,
				"--config-server-tls-certificate", configServerTLSConfig.Certificate,
				"--config-server-tls-private-key", configServerTLSConfig.PrivateKey,
				"--config-server-namespace", "config-server-ns",
				"--vars-store", "config-server://",
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = cmd.Opts.(*DeleteEnvOpts).VarsStore.List()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Listing of variables in config server is not supported"))
		})
	})

	Describe("alias-env command", func() {
		It("is passed global environment URL", func() {
			cmd, err := factory.New([]string{"alias-env", "-e", "env", "alias"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*AliasEnvOpts)
			Expect(opts.URL).To(Equal("env"))
		})

		It("is passed the global CA cert", func() {
			cmd, err := factory.New([]string{"alias-env", "--ca-cert", "BEGIN ca-cert", "alias"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*AliasEnvOpts)
			opts.CACert.FS = nil
			Expect(opts.CACert).To(Equal(CACertArg{Content: "BEGIN ca-cert"}))
		})
	})

	Describe("events command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"events", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*EventsOpts)
			Expect(opts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("vms command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"vms", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*VMsOpts)
			Expect(opts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("instances command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"instances", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*InstancesOpts)
			Expect(opts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("tasks command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"tasks", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*TasksOpts)
			Expect(opts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("task command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"task", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*TaskOpts)
			Expect(opts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("help command", func() {
		It("has a help command", func() {
			cmd, err := factory.New([]string{"help"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*MessageOpts)
			Expect(opts.Message).To(ContainSubstring("Usage:"))
			Expect(opts.Message).To(ContainSubstring("Application Options:"))
			Expect(opts.Message).To(ContainSubstring("Available commands:"))
		})
	})

	Describe("interpolate command", func() {
		It("is configured with config server flags for vars store", func() {
			err := fs.WriteFileString(fakeFilePath, "")
			Expect(err).ToNot(HaveOccurred())

			configServerTLSConfig := NewFakeTLSConfig()

			cmd, err := factory.New([]string{
				"interpolate", fakeFilePath,
				"--config-server-url", "config-server-url",
				"--config-server-tls-ca", configServerTLSConfig.CA,
				"--config-server-tls-certificate", configServerTLSConfig.Certificate,
				"--config-server-tls-private-key", configServerTLSConfig.PrivateKey,
				"--config-server-namespace", "config-server-ns",
				"--vars-store", "config-server://",
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = cmd.Opts.(*InterpolateOpts).VarsStore.List()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Listing of variables in config server is not supported"))
		})
	})

	Describe("help options", func() {
		It("has a help flag", func() {
			cmd, err := factory.New([]string{"--help"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*MessageOpts)
			Expect(opts.Message).To(ContainSubstring("Usage:"))
			Expect(opts.Message).To(ContainSubstring(
				"SSH into instance(s)                               https://bosh.io/docs/cli-v2#ssh"))
			Expect(opts.Message).To(ContainSubstring("Application Options:"))
			Expect(opts.Message).To(ContainSubstring("Available commands:"))
		})

		It("has a command help flag", func() {
			cmd, err := factory.New([]string{"ssh", "--help"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*MessageOpts)
			Expect(opts.Message).To(ContainSubstring("Usage:"))
			Expect(opts.Message).To(ContainSubstring("SSH into instance(s)\n\nhttps://bosh.io/docs/cli-v2#ssh"))
			Expect(opts.Message).To(ContainSubstring("Application Options:"))
			Expect(opts.Message).To(ContainSubstring("[ssh command options]"))
		})
	})

	Describe("version option", func() {
		It("has a version flag", func() {
			cmd, err := factory.New([]string{"--version"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*MessageOpts)
			Expect(opts.Message).To(Equal("version [DEV BUILD]\n"))
		})
	})

	Describe("global options", func() {
		clearNonGlobalOpts := func(boshOpts BoshOpts) BoshOpts {
			boshOpts.VersionOpt = nil // can't compare functions

			// fs is populated by factory.New
			boshOpts.CACertOpt.FS = nil
			boshOpts.ConfigServerFlags.TLSCA.FS = nil
			boshOpts.ConfigServerFlags.TLSCertificate.FS = nil
			boshOpts.ConfigServerFlags.TLSPrivateKey.FS = nil

			boshOpts.CreateEnv = CreateEnvOpts{}
			boshOpts.DeleteEnv = DeleteEnvOpts{}
			boshOpts.UploadRelease = UploadReleaseOpts{}
			boshOpts.ExportRelease = ExportReleaseOpts{}
			boshOpts.RunErrand = RunErrandOpts{}
			boshOpts.Logs = LogsOpts{}
			boshOpts.Interpolate = InterpolateOpts{}
			boshOpts.InitRelease = InitReleaseOpts{}
			boshOpts.ResetRelease = ResetReleaseOpts{}
			boshOpts.GenerateJob = GenerateJobOpts{}
			boshOpts.GeneratePackage = GeneratePackageOpts{}
			boshOpts.VendorPackage = VendorPackageOpts{}
			boshOpts.CreateRelease = CreateReleaseOpts{}
			boshOpts.FinalizeRelease = FinalizeReleaseOpts{}
			boshOpts.Blobs = BlobsOpts{}
			boshOpts.AddBlob = AddBlobOpts{}
			boshOpts.RemoveBlob = RemoveBlobOpts{}
			boshOpts.SyncBlobs = SyncBlobsOpts{}
			boshOpts.UploadBlobs = UploadBlobsOpts{}
			boshOpts.SSH = SSHOpts{}
			boshOpts.SCP = SCPOpts{}
			boshOpts.Deploy = DeployOpts{}
			boshOpts.UpdateCloudConfig = UpdateCloudConfigOpts{}
			boshOpts.UpdateCPIConfig = UpdateCPIConfigOpts{}
			boshOpts.UpdateRuntimeConfig = UpdateRuntimeConfigOpts{}
			boshOpts.VMs = VMsOpts{}
			boshOpts.Instances = InstancesOpts{}
			boshOpts.Config = ConfigOpts{}
			boshOpts.Configs = ConfigsOpts{}
			boshOpts.UpdateConfig = UpdateConfigOpts{}
			boshOpts.DeleteConfig = DeleteConfigOpts{}
			return boshOpts
		}

		It("has set of default options", func() {
			cmd, err := factory.New([]string{"locks"})
			Expect(err).ToNot(HaveOccurred())

			// Check against entire BoshOpts to avoid future missing assertions
			Expect(clearNonGlobalOpts(cmd.BoshOpts)).To(Equal(BoshOpts{
				ConfigPathOpt: "~/.bosh/config",
				Parallel:      5,
			}))
		})

		It("can set variety of options", func() {
			configServerTLSConfig := NewFakeTLSConfig()

			opts := []string{
				"--config", "config",
				"--environment", "env",
				"--ca-cert", "BEGIN ca-cert",
				"--client", "client",
				"--client-secret", "client-secret",
				"--deployment", "dep",
				"--json",
				"--tty",
				"--no-color",
				"--non-interactive",
				"--parallel", "123",
				"--config-server-url", "config-server-url",
				"--config-server-tls-ca", configServerTLSConfig.CA,
				"--config-server-tls-certificate", configServerTLSConfig.Certificate,
				"--config-server-tls-private-key", configServerTLSConfig.PrivateKey,
				"--config-server-namespace", "config-server-ns",
				"locks",
			}

			cmd, err := factory.New(opts)
			Expect(err).ToNot(HaveOccurred())

			Expect(clearNonGlobalOpts(cmd.BoshOpts)).To(Equal(BoshOpts{
				ConfigPathOpt:     "config",
				EnvironmentOpt:    "env",
				CACertOpt:         CACertArg{Content: "BEGIN ca-cert"},
				ClientOpt:         "client",
				ClientSecretOpt:   "client-secret",
				DeploymentOpt:     "dep",
				JSONOpt:           true,
				TTYOpt:            true,
				NoColorOpt:        true,
				NonInteractiveOpt: true,
				ConfigServerFlags: ConfigServerFlags{
					URL:            "config-server-url",
					TLSCA:          CACertArg{Content: configServerTLSConfig.CA},
					TLSCertificate: CACertArg{Content: configServerTLSConfig.Certificate},
					TLSPrivateKey:  CACertArg{Content: configServerTLSConfig.PrivateKey},
					Namespace:      "config-server-ns",
				},
				Parallel: 123,
			}))
		})

		It("errors when --user is set", func() {
			opts := []string{
				"--user", "foo",
				"--json",
				"--tty",
			}

			_, err := factory.New(opts)
			Expect(err).To(HaveOccurred())
		})

		It("errors when BOSH_USER is set", func() {
			os.Setenv("BOSH_USER", "bar")
			_, err := factory.New([]string{})
			Expect(err).To(HaveOccurred())
		})
	})
})

type fakeTLSConfig struct {
	CA          string
	Certificate string
	PrivateKey  string
}

func NewFakeTLSConfig() fakeTLSConfig {
	return fakeTLSConfig{
		CA:          "-----BEGIN CERTIFICATE-----\nMIIDFDCCAfygAwIBAgIRAOIUrZu8YXP+aC0Df+ERlyAwDQYJKoZIhvcNAQELBQAw\nMzEMMAoGA1UEBhMDVVNBMRYwFAYDVQQKEw1DbG91ZCBGb3VuZHJ5MQswCQYDVQQD\nEwJjYTAeFw0xODAxMzEwMjA0MDZaFw0xOTAxMzEwMjA0MDZaMDMxDDAKBgNVBAYT\nA1VTQTEWMBQGA1UEChMNQ2xvdWQgRm91bmRyeTELMAkGA1UEAxMCY2EwggEiMA0G\nCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDhe27Hrcc/bsrQK4Xp/jm5kebU9Mn3\nh4C3DMMJH3klIiAGCToxK9ygHdrXJzvdLFJlulw8qBYznRKmVMubW6MOeQBWhiB0\nuijT6ZGoiiwjQwvE6ASSVLSaE2byOnXHDPrdgXJg1BuBt7ZZ7VXp7bGTVjXetpz4\nZsfP8di4YcuC76PqxgVFNTyPpmPNoSY9DW5Up4tKLetiOpxRmb7+vGvPiqrCZ4Dx\npq8caKOYr5dkBYL2ndQbHO9zU05xPBttOxPwJkTmI2RRfw8U5D5Dj5QdA8Gwctua\n2mCjpjnfZ6qQcYMFCHLS9GWKxwfxM3ZhaXHbOPzdmWOWSP3dEtbGInQBAgMBAAGj\nIzAhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEB\nCwUAA4IBAQC2D2N71wABMVOzBeIvWF7FmTFUd0zQ428h6MzAXWoKrc0aIQyhVs0m\nJrp1ieFxk3r+nAMRQXI6/Fg2W6sKGGy+0MbgAuDAcn/cxmES3VJzydx0I3EK0MBF\nY0NUOq83F1EBMR0x/M9QdePNoLWqaB6LHX0cW/UpIZPvpdjmbxeBEs84KhT/pKcF\n11rdLLr7Gs4ndg0CI8DXqR5SB+cxo2FfQ+TzATVtkRN7bNXYm3kWbGAfxThs6zBs\nmua9ksakqsOViqQFruxCUeHMxCZ+Jr5LYsjwGS9h6DDVU3rG+EM1v5+rlzyrV/E6\n028qgNmb9GC86sUMJP63Pgnor+/emp9N\n-----END CERTIFICATE-----",
		Certificate: "-----BEGIN CERTIFICATE-----\nMIIDSjCCAjKgAwIBAgIQWyNlE989qt4Jq5e3uPnEvzANBgkqhkiG9w0BAQsFADAz\nMQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxCzAJBgNVBAMT\nAmNhMB4XDTE4MDEzMTAyMDQwNloXDTE5MDEzMTAyMDQwNlowPDEMMAoGA1UEBhMD\nVVNBMRYwFAYDVQQKEw1DbG91ZCBGb3VuZHJ5MRQwEgYDVQQDEwtzZXJ2ZXItY2Vy\ndDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALZCZGt0DlFNSiRCZU83\nefftOa+EaXmkrMMr1chlb4rdSE5ft3G6Jw9cmdMDB4ZrA3pbQo6ENcFQn1AI6h/2\nC8q2XASUzzO7vDo/oYfsK4YGXXCgVEKB+aQagKDBhPCFW2m6SaoaHjxfZtS4TVBL\nepr7SwZfFBxFKLlInTOdy+i/TARFNf+xszrffVlhRTbTozCwI0wrURQeXPU0V5kh\nM4vOyDrN0/K9GmgO9dNC5X7T1JS2BQ7vtceAf7eSCyiuP3Zi8Gja9/51l3JKAXA7\nceY0+r0U+c4yd2XhaY0gHPYzJegzFilF+bC311heMJpQYm5KdAmeHMOtWW1m2zeU\nxLECAwEAAaNRME8wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMB\nMAwGA1UdEwEB/wQCMAAwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMA0GCSqG\nSIb3DQEBCwUAA4IBAQANsC2yzCvdQFW0Lr7iVNr44O4J3HLVdMuMGoNoYBGGk7+k\njiExDkY1BNvXtYxGTQo8x0a9i/DzrT1qTsxYQbmSUa35vh08bMhht6aRr4G5LkEq\nneHgJF3oo0G7sEu06dokKZLWkpHtHU9s2csXrLWYCIcejyWhkJGlKTN6WYgum3dS\nWjMSq20Hn+SW/PcUTBpB+gJDKnz2XCX9Hu0elVJOIv+DsRBNlrU0LUyF79Dbnl6j\n/nW3vZF9r7+LfYgLdLB661hvyv5F2e97ic7mmekG7nRop5j074vNPCj8yMdADqVn\nue3UUC2R6+gWoOxY0HVhVlGX9vVJap+/1O7rHBLr\n-----END CERTIFICATE-----",
		PrivateKey:  "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAtkJka3QOUU1KJEJlTzd59+05r4RpeaSswyvVyGVvit1ITl+3\ncbonD1yZ0wMHhmsDeltCjoQ1wVCfUAjqH/YLyrZcBJTPM7u8Oj+hh+wrhgZdcKBU\nQoH5pBqAoMGE8IVbabpJqhoePF9m1LhNUEt6mvtLBl8UHEUouUidM53L6L9MBEU1\n/7GzOt99WWFFNtOjMLAjTCtRFB5c9TRXmSEzi87IOs3T8r0aaA7100LlftPUlLYF\nDu+1x4B/t5ILKK4/dmLwaNr3/nWXckoBcDtx5jT6vRT5zjJ3ZeFpjSAc9jMl6DMW\nKUX5sLfXWF4wmlBibkp0CZ4cw61ZbWbbN5TEsQIDAQABAoIBADL52tBbA24l6ei+\nUUuYvppjVVEL/dwx/MgRyJdmF46FWaXiC5LZd/dJ9RQZss8buztLrw/hVo+dFxHx\njFooHSAzZQU7AcD8bybziSBVI882lIfdr/NyGvqVFwjfV2lWQz0NB3F2IKLOJBq2\n+ZjNo5sZUeCUUzGc/kjkUGORbOjJr4OBkZM2hRqtdDbdaPK7OSiSFEpCwTuTLbhD\nCnP6ay+WWGADcgUSeYKrREohnjAhL8VuAOiQcveU4gmPfdMDZcq1cSdyEEO+NHs2\nFXLBKfdU42C61PRDfa5NNd1suTLBYAZWcr8AEnEvUEp/BeGhEBAdAlgVRHkjKs3G\nuD4GDGECgYEA03EQzRUbog6S2u8cIplUc5bJ9Wj+b+zmix22ve3FIxLTGfPb3cLL\nQe3JIkxLICYP7ZO1MaYy34+30lbPUbogkEzLP7fpt/qUk938CwLFL53ikOiWVA7k\np4hAC2G4Q20qP2biH6G4/h4eRhuXiJALu54GLtjMk2v9bQM2biYoWQUCgYEA3Kr8\n68vjIMU/xemllpafgVWICQMPOPCh6qdL9Fn7Atdt0BkjfDo3h1iaedEhKYQDZNvF\nxaiScSiByk9IFPyvIs0JEcywrKy5NQ4/0dMSQEOzfxBg9bPKI7/etHi9/ALW+XKB\naRZ4GNsr6YjC9szcE8LfQrCAsUe3BOX2zSUcnL0CgYAbiH+dlQASLD+nTrelMb4z\nhxEpadCoFns25lmjhdDD7nGa0Yxx5im9ng8w7ipiN1KfpzpTCsdZIUfYlgFNLSWM\nZNOaqoI+uNycHK3zaRrwRmj4YbEhpQbVYgKk+Mab0R1NQEJ1yANk49shWfpzh/5f\nIgbAFu8cy1Um2uI9ma5rWQKBgQDBdWankuhdIpD2ghCaJRNR4BqTTAtccBqEDoeY\nggp+Q0AS4PcrQh7MmfFUOvRH4WTYV5Tb5R399vVS2I7pV15ztC3vXPTHbeYxjXyG\nB/ZIQRJso39d6XGeReiJcBGfjx3JM4ohB4HiyMOGyk+i75dB++agIP2ybp0VvkbR\nM2gSQQKBgEm8KxlNQoI8FsV5I+9maL9ArFwUQX7Ck+5Z0/ldCviZdSAwDThY7Zng\nbY7my03T1A+gtVsgc/u8hsWuQh6wFPsB20wZoI0zDqcHnvmjgUUyQv18qhsAxiuO\nq3/JxAGV043+WleFaD9jGgBgPeqR3Aoch0U1AqDHDat5jVg2KLLT\n-----END RSA PRIVATE KEY-----",
	}
}
