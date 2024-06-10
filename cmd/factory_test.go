package cmd_test

import (
	"fmt"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

// This placeholder is used for replacing arguments in the table test with the
// temporary file created in the BeforeEach
const filePlaceholder = "replace-me"

var _ = Describe("Factory", func() {
	var (
		fs      boshsys.FileSystem
		factory cmd.Factory
		tmpFile string
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = boshsys.NewOsFileSystemWithStrictTempRoot(boshlog.NewLogger(boshlog.LevelNone))

		f, err := os.CreateTemp("", "file")
		Expect(err).NotTo(HaveOccurred())

		tmpFile = f.Name()

		ui := boshui.NewConfUI(logger)
		defer ui.Flush()

		deps := cmd.NewBasicDeps(ui, logger)
		deps.FS = fs

		factory = cmd.NewFactory(deps)
	})

	Context("extra args and flags", func() {
		DescribeTable("extra args and flags", func(cmd string, args []string) {
			for i, arg := range args {
				if arg == filePlaceholder {
					args[i] = tmpFile
				}
			}
			cmdWithArgs := append([]string{cmd}, args...)
			cmdWithArgs = append(cmdWithArgs, "extra", "args")

			_, err := factory.New(cmdWithArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not support extra arguments: extra, args"))
		},
			Entry("help", "help", []string{}),
			Entry("add-blob", "add-blob", []string{filePlaceholder, "directory"}),
			Entry("attach-disk", "attach-disk", []string{"instance/abad1dea", "disk-cid-123"}),
			Entry("blobs", "blobs", []string{}),
			Entry("interpolate", "interpolate", []string{filePlaceholder}),
			Entry("cancel-task", "cancel-task", []string{"1234"}),
			Entry("clean-up", "clean-up", []string{}),
			Entry("cloud-check", "cloud-check", []string{}),
			Entry("cloud-config", "cloud-config", []string{}),
			Entry("create-env", "create-env", []string{filePlaceholder}),
			Entry("sha2ify-release", "sha2ify-release", []string{filePlaceholder, filePlaceholder}),
			Entry("create-release", "create-release", []string{filePlaceholder}),
			Entry("delete-deployment", "delete-deployment", []string{}),
			Entry("delete-disk", "delete-disk", []string{"cid"}),
			Entry("delete-env", "delete-env", []string{filePlaceholder}),
			Entry("delete-release", "delete-release", []string{"release-version"}),
			Entry("delete-snapshot", "delete-snapshot", []string{"cid"}),
			Entry("delete-snapshots", "delete-snapshots", []string{}),
			Entry("delete-stemcell", "delete-stemcell", []string{"name/version"}),
			Entry("delete-vm", "delete-vm", []string{"cid"}),
			Entry("deploy", "deploy", []string{filePlaceholder}),
			Entry("deployment", "deployment", []string{}),
			Entry("deployments", "deployments", []string{}),
			Entry("disks", "disks", []string{}),
			Entry("alias-env", "alias-env", []string{"alias"}),
			Entry("environment", "environment", []string{}),
			Entry("environments", "environments", []string{}),
			Entry("errands", "errands", []string{}),
			Entry("events", "events", []string{}),
			Entry("export-release", "export-release", []string{"release/version", "os/version"}),
			Entry("finalize-release", "finalize-release", []string{filePlaceholder}),
			Entry("generate-job", "generate-job", []string{filePlaceholder}),
			Entry("generate-package", "generate-package", []string{filePlaceholder}),
			Entry("init-release", "init-release", []string{}),
			Entry("inspect-release", "inspect-release", []string{"name/version"}),
			Entry("inspect-local-release", "inspect-local-release", []string{filePlaceholder}),
			Entry("inspect-local-stemcell", "inspect-local-stemcell", []string{filePlaceholder}),
			Entry("instances", "instances", []string{}),
			Entry("locks", "locks", []string{}),
			Entry("log-in", "log-in", []string{}),
			Entry("log-out", "log-out", []string{}),
			Entry("logs", "logs", []string{"slug"}),
			Entry("manifest", "manifest", []string{}),
			Entry("recreate", "recreate", []string{"slug"}),
			Entry("releases", "releases", []string{}),
			Entry("remove-blob", "remove-blob", []string{filePlaceholder}),
			Entry("reset-release", "reset-release", []string{}),
			Entry("restart", "restart", []string{"slug"}),
			Entry("run-errand", "run-errand", []string{"name"}),
			Entry("runtime-config", "runtime-config", []string{}),
			Entry("snapshots", "snapshots", []string{"group/id"}),
			Entry("start", "start", []string{"slug"}),
			Entry("stemcells", "stemcells", []string{}),
			Entry("stop", "stop", []string{"slug"}),
			Entry("sync-blobs", "sync-blobs", []string{}),
			Entry("take-snapshot", "take-snapshot", []string{"group/id"}),
			Entry("task", "task", []string{"1234"}),
			Entry("tasks", "tasks", []string{}),
			Entry("update-cloud-config", "update-cloud-config", []string{filePlaceholder}),
			Entry("update-resurrection", "update-resurrection", []string{"off"}),
			Entry("update-runtime-config", "update-runtime-config", []string{filePlaceholder}),
			Entry("upload-blobs", "upload-blobs", []string{}),
			Entry("upload-release", "upload-release", []string{filePlaceholder}),
			Entry("upload-stemcell", "upload-stemcell", []string{filePlaceholder}),
			Entry("vms", "vms", []string{}),
			Entry("curl", "curl", []string{"/"}),
		)

		Describe("ssh", func() {
			It("uses all remaining arguments as a command", func() {
				cmd, err := factory.New([]string{"ssh", "group", "cmd", "extra", "args"})
				Expect(err).ToNot(HaveOccurred())

				sshOpts := cmd.Opts.(*opts.SSHOpts)
				Expect(sshOpts.Command).To(Equal([]string{"cmd", "extra", "args"}))
			})

			It("uses all remaining arguments as a command even that look like flags", func() {
				cmd, err := factory.New([]string{"ssh", "group", "cmd", "extra", "args", "--", "--gw-disable"})
				Expect(err).ToNot(HaveOccurred())

				sshOpts := cmd.Opts.(*opts.SSHOpts)
				Expect(sshOpts.Command).To(Equal([]string{"cmd", "extra", "args", "--gw-disable"}))
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

			_, _, err = cmd.Opts.(*opts.SSHOpts).GatewayFlags.AsSSHOpts()
			Expect(err).ToNot(HaveOccurred())
		})

		It("scp command has configured gateway flags", func() {
			cmd, err := factory.New([]string{"scp", "group", "cmd", "extra", "args", "--", "--gw-disable"})
			Expect(err).ToNot(HaveOccurred())

			_, _, err = cmd.Opts.(*opts.SCPOpts).GatewayFlags.AsSSHOpts()
			Expect(err).ToNot(HaveOccurred())
		})

		It("logs -f command has configured gateway flags", func() {
			cmd, err := factory.New([]string{"logs", "-f", "cmd"})
			Expect(err).ToNot(HaveOccurred())

			_, _, err = cmd.Opts.(*opts.LogsOpts).GatewayFlags.AsSSHOpts()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("deploy command", func() {
		It("parses multiple skip-drain flags", func() {
			cmd, err := factory.New([]string{"deploy", "--skip-drain=job1", "--skip-drain=job2", tmpFile})
			Expect(err).ToNot(HaveOccurred())

			slug1, _ := boshdir.NewInstanceGroupOrInstanceSlugFromString("job1")
			slug2, _ := boshdir.NewInstanceGroupOrInstanceSlugFromString("job2")

			deployOpts := cmd.Opts.(*opts.DeployOpts)
			Expect(deployOpts.SkipDrain).To(Equal([]boshdir.SkipDrain{
				{Slug: slug1},
				{Slug: slug2},
			}))
		})

		It("errors when excluding = from --skip-drain", func() {
			f, err := os.CreateTemp("", "job1")
			Expect(err).NotTo(HaveOccurred())

			nonExistantPath := f.Name()
			Expect(os.RemoveAll(nonExistantPath)).To(Succeed())

			_, err = factory.New([]string{"deploy", "--skip-drain", nonExistantPath, tmpFile})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})

		It("defaults --skip-drain option value to all", func() {
			cmd, err := factory.New([]string{"deploy", "--skip-drain", tmpFile})
			Expect(err).ToNot(HaveOccurred())

			deployOpts := cmd.Opts.(*opts.DeployOpts)
			Expect(deployOpts.SkipDrain).To(Equal([]boshdir.SkipDrain{
				{All: true},
			}))
		})
	})

	Describe("create-env command (command that uses FileBytesArg)", func() {
		BeforeEach(func() {
			Expect(os.RemoveAll(tmpFile)).To(Succeed())
		})

		It("returns *nice error from FileBytesArg* error if it cannot read manifest", func() {
			_, err := factory.New([]string{"create-env", tmpFile})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("open %s: no such file or directory", tmpFile)))
		})
	})

	Describe("alias-env command", func() {
		It("is passed global environment URL", func() {
			cmd, err := factory.New([]string{"alias-env", "-e", "env", "alias"})
			Expect(err).ToNot(HaveOccurred())

			aliasEnvOpts := cmd.Opts.(*opts.AliasEnvOpts)
			Expect(aliasEnvOpts.URL).To(Equal("env"))
		})

		It("is passed the global CA cert", func() {
			cmd, err := factory.New([]string{"alias-env", "--ca-cert", "BEGIN ca-cert", "alias"})
			Expect(err).ToNot(HaveOccurred())

			aliasEnvOpts := cmd.Opts.(*opts.AliasEnvOpts)
			aliasEnvOpts.CACert.FS = nil
			Expect(aliasEnvOpts.CACert).To(Equal(opts.CACertArg{Content: "BEGIN ca-cert"}))
		})
	})

	Describe("events command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"events", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			eventsOpts := cmd.Opts.(*opts.EventsOpts)
			Expect(eventsOpts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("vms command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"vms", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			vMsOpts := cmd.Opts.(*opts.VMsOpts)
			Expect(vMsOpts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("instances command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"instances", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			instancesOpts := cmd.Opts.(*opts.InstancesOpts)
			Expect(instancesOpts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("tasks command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"tasks", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			tasksOpts := cmd.Opts.(*opts.TasksOpts)
			Expect(tasksOpts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("task command", func() {
		It("is passed the deployment flag", func() {
			cmd, err := factory.New([]string{"task", "--deployment", "deployment"})
			Expect(err).ToNot(HaveOccurred())

			taskOpts := cmd.Opts.(*opts.TaskOpts)
			Expect(taskOpts.Deployment).To(Equal("deployment"))
		})
	})

	Describe("help command", func() {
		It("has a help command", func() {
			cmd, err := factory.New([]string{"help"})
			Expect(err).ToNot(HaveOccurred())

			messageOpts := cmd.Opts.(*opts.MessageOpts)
			Expect(messageOpts.Message).To(ContainSubstring("Usage:"))
			Expect(messageOpts.Message).To(ContainSubstring("Application Options:"))
			Expect(messageOpts.Message).To(ContainSubstring("Available commands:"))
		})
	})

	Describe("help options", func() {
		It("has a help flag", func() {
			cmd, err := factory.New([]string{"--help"})
			Expect(err).ToNot(HaveOccurred())

			messageOpts := cmd.Opts.(*opts.MessageOpts)
			Expect(messageOpts.Message).To(ContainSubstring("Usage:"))
			Expect(messageOpts.Message).To(ContainSubstring(
				"SSH into instance(s)                               https://bosh.io/docs/cli-v2#ssh"))
			Expect(messageOpts.Message).To(ContainSubstring("Application Options:"))
			Expect(messageOpts.Message).To(ContainSubstring("Available commands:"))
		})

		It("has a command help flag", func() {
			cmd, err := factory.New([]string{"ssh", "--help"})
			Expect(err).ToNot(HaveOccurred())

			messageOpts := cmd.Opts.(*opts.MessageOpts)
			Expect(messageOpts.Message).To(ContainSubstring("Usage:"))
			Expect(messageOpts.Message).To(ContainSubstring("SSH into instance(s)\n\nhttps://bosh.io/docs/cli-v2#ssh"))
			Expect(messageOpts.Message).To(ContainSubstring("Application Options:"))
			Expect(messageOpts.Message).To(ContainSubstring("[ssh command options]"))
		})
	})

	Describe("version option", func() {
		It("has a version flag", func() {
			cmd, err := factory.New([]string{"--version"})
			Expect(err).ToNot(HaveOccurred())

			messageOpts := cmd.Opts.(*opts.MessageOpts)
			Expect(messageOpts.Message).To(Equal("version [DEV BUILD]\n"))
		})
	})

	Describe("global options", func() {
		clearNonGlobalOpts := func(boshOpts opts.BoshOpts) opts.BoshOpts {
			boshOpts.VersionOpt = nil   // can't compare functions
			boshOpts.CACertOpt.FS = nil // fs is populated by factory.New
			boshOpts.UploadRelease = opts.UploadReleaseOpts{}
			boshOpts.ExportRelease = opts.ExportReleaseOpts{}
			boshOpts.RunErrand = opts.RunErrandOpts{}
			boshOpts.Logs = opts.LogsOpts{}
			boshOpts.Interpolate = opts.InterpolateOpts{}
			boshOpts.InitRelease = opts.InitReleaseOpts{}
			boshOpts.ResetRelease = opts.ResetReleaseOpts{}
			boshOpts.GenerateJob = opts.GenerateJobOpts{}
			boshOpts.GeneratePackage = opts.GeneratePackageOpts{}
			boshOpts.VendorPackage = opts.VendorPackageOpts{}
			boshOpts.CreateRelease = opts.CreateReleaseOpts{}
			boshOpts.FinalizeRelease = opts.FinalizeReleaseOpts{}
			boshOpts.Blobs = opts.BlobsOpts{}
			boshOpts.AddBlob = opts.AddBlobOpts{}
			boshOpts.RemoveBlob = opts.RemoveBlobOpts{}
			boshOpts.SyncBlobs = opts.SyncBlobsOpts{}
			boshOpts.UploadBlobs = opts.UploadBlobsOpts{}
			boshOpts.Pcap = opts.PcapOpts{}
			boshOpts.SSH = opts.SSHOpts{}
			boshOpts.SCP = opts.SCPOpts{}
			boshOpts.Deploy = opts.DeployOpts{}
			boshOpts.UpdateRuntimeConfig = opts.UpdateRuntimeConfigOpts{}
			boshOpts.VMs = opts.VMsOpts{}
			boshOpts.Instances = opts.InstancesOpts{}
			boshOpts.Config = opts.ConfigOpts{}
			boshOpts.Configs = opts.ConfigsOpts{}
			boshOpts.UpdateConfig = opts.UpdateConfigOpts{}
			boshOpts.DeleteConfig = opts.DeleteConfigOpts{}
			boshOpts.Curl = opts.CurlOpts{}
			return boshOpts
		}

		It("has set of default options", func() {
			cmd, err := factory.New([]string{"locks"})
			Expect(err).ToNot(HaveOccurred())

			// Check against entire BoshOpts to avoid future missing assertions
			Expect(clearNonGlobalOpts(cmd.BoshOpts)).To(Equal(opts.BoshOpts{
				ConfigPathOpt: "~/.bosh/config",
				Parallel:      5,
			}))
		})

		It("can set variety of options", func() {
			optsArray := []string{
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
				"locks",
			}

			cmd, err := factory.New(optsArray)
			Expect(err).ToNot(HaveOccurred())

			Expect(clearNonGlobalOpts(cmd.BoshOpts)).To(Equal(opts.BoshOpts{
				ConfigPathOpt:     "config",
				EnvironmentOpt:    "env",
				CACertOpt:         opts.CACertArg{Content: "BEGIN ca-cert"},
				ClientOpt:         "client",
				ClientSecretOpt:   "client-secret",
				DeploymentOpt:     "dep",
				JSONOpt:           true,
				TTYOpt:            true,
				NoColorOpt:        true,
				NonInteractiveOpt: true,
				Parallel:          123,
			}))
		})

		It("errors when --user is set", func() {
			optsArray := []string{
				"--user", "foo",
				"--json",
				"--tty",
			}

			_, err := factory.New(optsArray)
			Expect(err).To(HaveOccurred())
		})
	})
})
