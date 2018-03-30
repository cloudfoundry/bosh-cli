package cmd_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/cmd/config/configfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteAliasCmd", func() {
	var (
		config  *configfakes.FakeConfig
		command *cmd.DeleteAliasCmd
		opts    cmd.DeleteAliasOpts
	)

	BeforeEach(func() {
		config = &configfakes.FakeConfig{}
		command = cmd.NewDeleteAliasCmd(config)
		opts = cmd.DeleteAliasOpts{Args: cmd.DeleteAliasArgs{Alias: "foo"}}
	})

	Describe("Run", func() {
		It("returns an error if deleting fails", func() {
			config.DeleteAliasReturns(nil, errors.New("cannot delete"))
			Expect(command.Run(opts)).NotTo(Succeed())
		})

		It("saves the updated config", func() {
			config.DeleteAliasReturns(config, nil)
			config.SaveReturns(nil)
			Expect(command.Run(opts)).To(Succeed())
			Expect(config.SaveCallCount()).To(Equal(1))
		})

		It("returns an error if it cannot save the config", func() {
			config.DeleteAliasReturns(config, nil)
			config.SaveReturns(errors.New("cannot save"))
			Expect(command.Run(opts)).NotTo(Succeed())
		})
	})
})
