package cmd_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/cmd/config/configfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnaliasEnvCmd", func() {
	var (
		config  *configfakes.FakeConfig
		command *cmd.UnaliasEnvCmd
		opts    cmd.UnaliasEnvOpts
	)

	BeforeEach(func() {
		config = &configfakes.FakeConfig{}
		command = cmd.NewUnaliasEnvCmd(config)
		opts = cmd.UnaliasEnvOpts{Args: cmd.UnaliasEnvArgs{Alias: "foo"}}
	})

	Describe("Run", func() {
		It("returns an error if deleting fails", func() {
			config.UnaliasEnvironmentReturns(nil, errors.New("cannot delete"))
			Expect(command.Run(opts)).NotTo(Succeed())
		})

		It("saves the updated config", func() {
			config.UnaliasEnvironmentReturns(config, nil)
			config.SaveReturns(nil)
			Expect(command.Run(opts)).To(Succeed())
			Expect(config.SaveCallCount()).To(Equal(1))
		})

		It("returns an error if it cannot save the config", func() {
			config.UnaliasEnvironmentReturns(config, nil)
			config.SaveReturns(errors.New("cannot save"))
			Expect(command.Run(opts)).NotTo(Succeed())
		})
	})
})
