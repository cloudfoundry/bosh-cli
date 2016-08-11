package cmd_test

import (
	boshcmd "github.com/cloudfoundry/bosh-init/cmd"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {
	Describe("extra arg handling", func() {
		var logger boshlog.Logger
		var ui *boshui.ConfUI
		var cmdFactory boshcmd.Factory

		BeforeEach(func() {
			logger = boshlog.NewLogger(boshlog.LevelNone)

			ui = boshui.NewConfUI(logger)
			defer ui.Flush()

			cmdFactory = boshcmd.NewFactory(boshcmd.NewBasicDeps(ui, logger))
		})

		AssertFailsWithExtraArguments := func(command string, args []string) func() {
			return func() {
				commandWithArgs := append([]string{command}, args...)
				err := cmdFactory.RunCommand(append(commandWithArgs, "more", "args", "thatarebad", "moreargs"))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Extra arguments are not supported for this command:"))
				Expect(err.Error()).To(ContainSubstring("thatarebad, moreargs"))
			}
		}

		Describe("create-env", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("create-env", []string{"/dev/null"}))
		})

		Describe("delete-env", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("delete-env", []string{"/dev/null"}))
		})

		Describe("environment", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("environment", []string{}))
		})

		Describe("environments", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("environments", []string{}))
		})

		Describe("log-in", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("log-in", []string{}))
		})

		Describe("log-out", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("log-out", []string{}))
		})

		Describe("task", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("task", []string{"1234"}))
		})

		Describe("tasks", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("tasks", []string{}))
		})

		Describe("cancel-task", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("cancel-task", []string{"1234"}))
		})

		Describe("locks", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("locks", []string{}))
		})

		Describe("clean-up", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("clean-up", []string{}))
		})

		Describe("build-manifest", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("build-manifest", []string{"/dev/null"}))
		})

		Describe("cloud-config", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("cloud-config", []string{}))
		})

		Describe("update-cloud-config", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("update-cloud-config", []string{"/dev/null"}))
		})

		Describe("runtime-config", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("runtime-config", []string{}))
		})

		Describe("update-runtime-config", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("update-runtime-config", []string{"/dev/null"}))
		})

		Describe("deployments", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("deployments", []string{}))
		})

		Describe("delete-deployment", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("delete-deployment", []string{}))
		})

		Describe("deploy", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("deploy", []string{"/dev/null"}))
		})

		Describe("manifest", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("manifest", []string{}))
		})

		Describe("stemcells", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("stemcells", []string{}))
		})

		Describe("upload-stemcell", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("upload-stemcell", []string{"/dev/null"}))
		})

		Describe("delete-stemcell", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("delete-stemcell", []string{"name/version"}))
		})

		Describe("deployment", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("deployment", []string{"/dev/null"}))
		})

		Describe("releases", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("releases", []string{}))
		})

		Describe("upload-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("upload-release", []string{"/dev/null"}))
		})

		Describe("export-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("export-release", []string{"release/version", "os/version"}))
		})

		Describe("inspect-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("inspect-release", []string{"name/version"}))
		})

		Describe("delete-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("delete-release", []string{"release-version"}))
		})

		Describe("errands", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("errands", []string{}))
		})

		Describe("run-errand", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("run-errand", []string{"name"}))
		})

		Describe("disks", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("disks", []string{}))
		})

		Describe("delete-disk", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("delete-disk", []string{"cid"}))
		})

		Describe("snapshots", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("snapshots", []string{"group/id"}))
		})

		Describe("take-snapshot", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("take-snapshot", []string{"group/id"}))
		})

		Describe("delete-snapshot", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("delete-snapshot", []string{"cid"}))
		})

		Describe("delete-snapshots", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("delete-snapshots", []string{"target"}))
		})

		Describe("instances", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("instances", []string{}))
		})

		Describe("vms", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("vms", []string{}))
		})

		Describe("vm-resurrection", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("vm-resurrection", []string{"off"}))
		})

		Describe("cloud-check", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("cloud-check", []string{}))
		})

		Describe("logs", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("logs", []string{"slug"}))
		})

		Describe("start", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("start", []string{"slug"}))
		})

		Describe("stop", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("stop", []string{"slug"}))
		})

		Describe("restart", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("restart", []string{"slug"}))
		})

		Describe("recreate", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("recreate", []string{"slug"}))
		})

		Describe("init-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("init-release", []string{"/dev/null"}))
		})

		Describe("reset-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("reset-release", []string{}))
		})

		Describe("generate-job", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("generate-job", []string{"/dev/null"}))
		})

		Describe("generate-package", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("generate-package", []string{"/dev/null"}))
		})

		Describe("create-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("create-release", []string{"/dev/null"}))
		})

		Describe("finalize-release", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("finalize-release", []string{"/dev/null"}))
		})

		Describe("blobs", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("blobs", []string{}))
		})

		Describe("add-blob", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("add-blob", []string{"/dev/null"}))
		})

		Describe("remove-blob", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("remove-blob", []string{"/dev/null"}))
		})

		Describe("sync-blobs", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("sync-blobs", []string{"/dev/null"}))
		})

		Describe("upload-blobs", func() {
			It("fails with extra args", AssertFailsWithExtraArguments("upload-blobs", []string{}))
		})
	})
})
