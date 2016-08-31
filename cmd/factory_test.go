package cmd_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshui "github.com/cloudfoundry/bosh-init/ui"
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
			"create-env":            []string{"/file"},
			"delete-env":            []string{"/file"},
			"environment":           []string{"url", "alias"},
			"environments":          []string{},
			"log-in":                []string{},
			"log-out":               []string{},
			"task":                  []string{"1234"},
			"tasks":                 []string{},
			"cancel-task":           []string{"1234"},
			"locks":                 []string{},
			"clean-up":              []string{},
			"build-manifest":        []string{"/file"},
			"cloud-config":          []string{},
			"update-cloud-config":   []string{"/file"},
			"runtime-config":        []string{},
			"update-runtime-config": []string{"/file"},
			"deployments":           []string{},
			"delete-deployment":     []string{},
			"deploy":                []string{"/file"},
			"manifest":              []string{},
			"events":                []string{},
			"stemcells":             []string{},
			"upload-stemcell":       []string{"/file"},
			"delete-stemcell":       []string{"name/version"},
			"deployment":            []string{"/file"},
			"releases":              []string{},
			"upload-release":        []string{"/file"},
			"export-release":        []string{"release/version", "os/version"},
			"inspect-release":       []string{"name/version"},
			"delete-release":        []string{"release-version"},
			"errands":               []string{},
			"run-errand":            []string{"name"},
			"disks":                 []string{},
			"delete-disk":           []string{"cid"},
			"snapshots":             []string{"group/id"},
			"take-snapshot":         []string{"group/id"},
			"delete-snapshot":       []string{"cid"},
			"delete-snapshots":      []string{},
			"instances":             []string{},
			"vms":                   []string{},
			"vm-resurrection":       []string{"off"},
			"cloud-check":           []string{},
			"logs":                  []string{"slug"},
			"start":                 []string{"slug"},
			"stop":                  []string{"slug"},
			"restart":               []string{"slug"},
			"recreate":              []string{"slug"},
			"init-release":          []string{},
			"reset-release":         []string{},
			"generate-job":          []string{"/file"},
			"generate-package":      []string{"/file"},
			"create-release":        []string{"/file"},
			"finalize-release":      []string{"/file"},
			"blobs":                 []string{},
			"add-blob":              []string{"/file", "directory"},
			"remove-blob":           []string{"/file"},
			"sync-blobs":            []string{},
			"upload-blobs":          []string{},
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
		It("returns an error if it cannot read manifest", func() {
			fs.ReadFileError = errors.New("fake-err")

			_, err := factory.New([]string{"create-env", "manifest.yml"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("the required argument `PATH` was not provided"))
		})
	})

	Describe("environment command", func() {
		It("is passed the global CA cert", func() {
			cmd, err := factory.New([]string{"environment", "--ca-cert", "ca-cert"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*EnvironmentOpts)
			Expect(opts.CACert).To(Equal("ca-cert"))
		})
	})

	Describe("events command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"events", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			opts := cmd.Opts.(*EventsOpts)
			Expect(*opts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("help options", func() {
		It("has a help flag", func() {
			_, err := factory.New([]string{"--help"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Usage:"))
			Expect(err.Error()).To(ContainSubstring("Application Options:"))
			Expect(err.Error()).To(ContainSubstring("Available commands:"))
		})

		It("has a command help flag", func() {
			_, err := factory.New([]string{"ssh", "--help"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Usage:"))
			Expect(err.Error()).To(ContainSubstring("Application Options:"))
			Expect(err.Error()).To(ContainSubstring("[ssh command options]"))
		})
	})

	Describe("version option", func() {
		It("has a version flag", func() {
			_, err := factory.New([]string{"--version"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("version [DEV BUILD]"))
		})
	})
})
