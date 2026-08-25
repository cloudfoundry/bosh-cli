package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmdconf "github.com/cloudfoundry/bosh-cli/v7/cmd/config/configfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("LogOutCmd", func() {
	var (
		config  *fakecmdconf.FakeConfig
		testUI  *testui.Ui
		command cmd.LogOutCmd
	)

	BeforeEach(func() {
		config = &fakecmdconf.FakeConfig{}
		testUI = &testui.Ui{}
		command = cmd.NewLogOutCmd("environment", config, testUI)
	})

	Describe("Run", func() {
		var (
			updatedConfig *fakecmdconf.FakeConfig
		)

		BeforeEach(func() {
			updatedConfig = &fakecmdconf.FakeConfig{}
			config.UnsetCredentialsReturns(updatedConfig)
		})

		act := func() error { return command.Run() }

		It("unsets credentials for the specific environment and saves config", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.UnsetCredentialsCallCount()).To(Equal(1))
			Expect(config.UnsetCredentialsArgsForCall(0)).To(Equal("environment"))

			Expect(updatedConfig.SaveCallCount()).To(Equal(1))

			Expect(testUI.Said).To(Equal([]string{"Logged out from 'environment'"}))
		})

		It("returns error if saving config failed", func() {
			updatedConfig.SaveReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(testUI.Said).To(BeEmpty())
		})

		It("returns error if environment is empty", func() {
			command = cmd.NewLogOutCmd("", config, testUI)
			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty Director URL"))
		})
	})
})
