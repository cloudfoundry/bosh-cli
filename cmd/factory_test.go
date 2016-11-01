package cmd_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

var _ = Describe("Factory", func() {
	var (
		fs      *fakesys.FakeFileSystem
		factory Factory
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()

		ui := boshui.NewConfUI(logger)
		defer ui.Flush()

		deps := NewBasicDeps(ui, logger)
		deps.FS = fs

		factory = NewFactory(deps)
	})

	Describe("unknown commands, args and flags", func() {
		BeforeEach(func() {
			err := fs.WriteFileString("/file", "")
			Expect(err).ToNot(HaveOccurred())
		})

		cmds := map[string][]string{
			"add-blob":              []string{"/file", "directory"},
			"blobs":                 []string{},
			"build-manifest":        []string{"/file"},
			"cancel-task":           []string{"1234"},
			"clean-up":              []string{},
			"cloud-check":           []string{},
			"cloud-config":          []string{},
			"create-env":            []string{"/file"},
			"create-release":        []string{"/file"},
			"delete-deployment":     []string{},
			"delete-disk":           []string{"cid"},
			"delete-env":            []string{"/file"},
			"delete-release":        []string{"release-version"},
			"delete-snapshot":       []string{"cid"},
			"delete-snapshots":      []string{},
			"delete-stemcell":       []string{"name/version"},
			"delete-vm":             []string{"cid"},
			"deploy":                []string{"/file"},
			"deployment":            []string{},
			"deployments":           []string{},
			"disks":                 []string{},
			"alias-env":             []string{"alias"},
			"environment":           []string{},
			"environments":          []string{},
			"errands":               []string{},
			"events":                []string{},
			"export-release":        []string{"release/version", "os/version"},
			"finalize-release":      []string{"/file"},
			"generate-job":          []string{"/file"},
			"generate-package":      []string{"/file"},
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
			"remove-blob":           []string{"/file"},
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
			"update-cloud-config":   []string{"/file"},
			"update-resurrection":   []string{"off"},
			"update-runtime-config": []string{"/file"},
			"upload-blobs":          []string{},
			"upload-release":        []string{"/file"},
			"upload-stemcell":       []string{"/file"},
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

	Describe("create-env command (command that uses FileBytesArg)", func() {
		It("returns *nice error from FileBytesArg* error if it cannot read manifest", func() {
			fs.ReadFileError = errors.New("fake-err")

			_, err := factory.New([]string{"create-env", "manifest.yml"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("File not found: 'manifest.yml'"))
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
			cmd, err := factory.New([]string{"alias-env", "--ca-cert", "ca-cert", "alias"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*AliasEnvOpts)
			Expect(opts.CACert).To(Equal("ca-cert"))
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

	Describe("tasks command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"tasks", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*TasksOpts)
			Expect(opts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("help options", func() {
		It("has a help flag", func() {
			cmd, err := factory.New([]string{"--help"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*MessageOpts)
			Expect(opts.Message).To(ContainSubstring("Usage:"))
			Expect(opts.Message).To(ContainSubstring("Application Options:"))
			Expect(opts.Message).To(ContainSubstring("Available commands:"))
		})

		It("has a command help flag", func() {
			cmd, err := factory.New([]string{"ssh", "--help"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*MessageOpts)
			Expect(opts.Message).To(ContainSubstring("Usage:"))
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
			boshOpts.UploadRelease = UploadReleaseOpts{}
			boshOpts.ExportRelease = ExportReleaseOpts{}
			boshOpts.RunErrand = RunErrandOpts{}
			boshOpts.Logs = LogsOpts{}
			boshOpts.InitRelease = InitReleaseOpts{}
			boshOpts.ResetRelease = ResetReleaseOpts{}
			boshOpts.GenerateJob = GenerateJobOpts{}
			boshOpts.GeneratePackage = GeneratePackageOpts{}
			boshOpts.CreateRelease = CreateReleaseOpts{}
			boshOpts.FinalizeRelease = FinalizeReleaseOpts{}
			boshOpts.Blobs = BlobsOpts{}
			boshOpts.AddBlob = AddBlobOpts{}
			boshOpts.RemoveBlob = RemoveBlobOpts{}
			boshOpts.SyncBlobs = SyncBlobsOpts{}
			boshOpts.UploadBlobs = UploadBlobsOpts{}
			return boshOpts
		}

		It("has set of default options", func() {
			cmd, err := factory.New([]string{"locks"})
			Expect(err).ToNot(HaveOccurred())

			// Check against entire BoshOpts to avoid future missing assertions
			Expect(clearNonGlobalOpts(cmd.BoshOpts)).To(Equal(BoshOpts{
				ConfigPathOpt: "~/.bosh/config",
			}))
		})

		It("can set variety of options", func() {
			opts := []string{
				"--config", "config",
				"--environment", "env",
				"--ca-cert", "ca-cert",
				"--user", "user",
				"--password", "password",
				"--uaa-client", "uaa-client",
				"--uaa-client-secret", "uaa-client-secret",
				"--deployment", "dep",
				"--json",
				"--tty",
				"--no-color",
				"--non-interactive",
				"locks",
			}

			cmd, err := factory.New(opts)
			Expect(err).ToNot(HaveOccurred())

			Expect(clearNonGlobalOpts(cmd.BoshOpts)).To(Equal(BoshOpts{
				ConfigPathOpt:      "config",
				EnvironmentOpt:     "env",
				CACertOpt:          "ca-cert",
				UsernameOpt:        "user",
				PasswordOpt:        "password",
				UAAClientOpt:       "uaa-client",
				UAAClientSecretOpt: "uaa-client-secret",
				DeploymentOpt:      "dep",
				JSONOpt:            true,
				TTYOpt:             true,
				NoColorOpt:         true,
				NonInteractiveOpt:  true,
			}))
		})
	})
})
