package cmd_test

import (
	"os"
	"os/user"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

var _ = Describe("cmd.Factory", func() {
	var (
		factory    Factory
		logger     boshlog.Logger
		filesystem boshsys.FileSystem
		ui         bmui.UI
		usr        *user.User
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelDebug)
		factory = NewFactory(logger)
		filesystem = boshsys.NewOsFileSystem(logger)
		ui = bmui.NewDefaultUI(os.Stdout, os.Stderr)
		usr, _ = user.Current()
	})

	It("creates a new factory", func() {
		Expect(factory).ToNot(BeNil())
	})

	Context("passing correct command name", func() {
		It("has deployment command", func() {
			cmd, err := factory.CreateCommand("deployment")
			Expect(err).ToNot(HaveOccurred())
			Expect(cmd).To(Equal(NewDeploymentCmd(ui, usr.HomeDir, filesystem)))
		})
	})

	Context("invalid command name", func() {
		It("returns error on invalid command name", func() {
			_, err := factory.CreateCommand("bogus-cmd-name")
			Expect(err).To(HaveOccurred())
		})
	})
})
