package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	fakecmdconf "github.com/cloudfoundry/bosh-init/cmd/config/fakes"
	fakecmd "github.com/cloudfoundry/bosh-init/cmd/fakes"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		sessions map[cmdconf.Config]*fakecmd.FakeSession
		config   *fakecmdconf.FakeConfig
		ui       *fakeui.FakeUI
		command  DeploymentCmd
	)

	BeforeEach(func() {
		sessions = map[cmdconf.Config]*fakecmd.FakeSession{}
		sessionFactory := func(config cmdconf.Config) Session {
			return sessions[config]
		}
		config = &fakecmdconf.FakeConfig{}
		ui = &fakeui.FakeUI{}
		command = NewDeploymentCmd(sessionFactory, config, ui)
	})

	Describe("Run", func() {
		var (
			opts           DeploymentOpts
			initialSession *fakecmd.FakeSession
			updatedSession *fakecmd.FakeSession
			updatedConfig  *fakecmdconf.FakeConfig
			deployment     *fakedir.FakeDeployment
		)

		BeforeEach(func() {
			opts = DeploymentOpts{}

			initialSession = &fakecmd.FakeSession{}
			sessions[config] = initialSession

			initialSession.TargetReturns("target-url")
		})

		act := func() error { return command.Run(opts) }

		Context("when deployment name/path arg is given", func() {
			BeforeEach(func() {
				opts.Args.NameOrPath = "deployment-name"

				updatedConfig = &fakecmdconf.FakeConfig{}
				config.SetDeploymentReturns(updatedConfig)

				updatedSession = &fakecmd.FakeSession{}
				sessions[updatedConfig] = updatedSession

				deployment = &fakedir.FakeDeployment{
					NameStub: func() string { return "deployment-name" },
				}
				updatedSession.DeploymentReturns(deployment, nil)
			})

			Context("when deployment is successfully determined", func() {
				It("saves config and sets current deployment", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(config.SetDeploymentCallCount()).To(Equal(1))

					target, nameOrPath := config.SetDeploymentArgsForCall(0)
					Expect(target).To(Equal("target-url"))
					Expect(nameOrPath).To(Equal("deployment-name"))

					Expect(updatedConfig.SaveCallCount()).To(Equal(1))

					Expect(ui.Said).To(Equal([]string{"Deployment set to 'deployment-name'"}))
				})
			})

			Context("when deployment cannot be determined", func() {
				BeforeEach(func() {
					updatedSession.DeploymentReturns(nil, errors.New("fake-err"))
				})

				It("returns an error and does not save config", func() {
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(updatedConfig.SaveCallCount()).To(Equal(0))

					Expect(ui.Said).To(BeEmpty())
				})
			})

			Context("when saving config fails", func() {
				BeforeEach(func() {
					updatedConfig.SaveReturns(errors.New("fake-err"))
				})

				It("returns an error", func() {
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(ui.Said).To(BeEmpty())
				})
			})
		})

		Context("when no args are given", func() {
			Context("when current deployment can be determined", func() {
				BeforeEach(func() {
					deployment = &fakedir.FakeDeployment{
						NameStub: func() string { return "deployment-name" },
					}
					initialSession.DeploymentReturns(deployment, nil)
				})

				It("shows current deployment name", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Said).To(Equal([]string{"Current deployment is 'deployment-name'"}))
				})
			})

			Context("when current deployment cannot be determined", func() {
				BeforeEach(func() {
					initialSession.DeploymentReturns(nil, errors.New("fake-err"))
				})

				It("returns an error", func() {
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(ui.Said).To(BeEmpty())
				})
			})
		})
	})
})
