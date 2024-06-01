package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/config/configfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

var _ = Describe("UnaliasEnvCmd", func() {
	var (
		config         *configfakes.FakeConfig
		command        *cmd.UnaliasEnvCmd
		unaliasEnvOpts opts.UnaliasEnvOpts
	)

	BeforeEach(func() {
		config = &configfakes.FakeConfig{}
		command = cmd.NewUnaliasEnvCmd(config)
		unaliasEnvOpts = opts.UnaliasEnvOpts{Args: opts.UnaliasEnvArgs{Alias: "foo"}}
	})

	Describe("Run", func() {
		It("returns an error if deleting fails", func() {
			config.UnaliasEnvironmentReturns(nil, errors.New("cannot delete"))
			Expect(command.Run(unaliasEnvOpts)).NotTo(Succeed())
		})

		It("saves the updated config", func() {
			config.UnaliasEnvironmentReturns(config, nil)
			config.SaveReturns(nil)
			Expect(command.Run(unaliasEnvOpts)).To(Succeed())
			Expect(config.SaveCallCount()).To(Equal(1))
		})

		It("returns an error if it cannot save the config", func() {
			config.UnaliasEnvironmentReturns(config, nil)
			config.SaveReturns(errors.New("cannot save"))
			Expect(command.Run(unaliasEnvOpts)).NotTo(Succeed())
		})
	})
})
