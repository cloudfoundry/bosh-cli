package cmd_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	cmdconf "github.com/cloudfoundry/bosh-cli/v7/cmd/config"
	fakecmdconf "github.com/cloudfoundry/bosh-cli/v7/cmd/config/configfakes"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("BasicLoginStrategy", func() {
	var (
		sessions map[cmdconf.Config]*fakecmd.FakeSession
		config   *fakecmdconf.FakeConfig
		testUI   *testui.Ui
		strategy cmd.BasicLoginStrategy
	)

	BeforeEach(func() {
		sessions = map[cmdconf.Config]*fakecmd.FakeSession{}
		sessionFactory := func(config cmdconf.Config) cmd.Session {
			return sessions[config]
		}
		config = &fakecmdconf.FakeConfig{}
		testUI = &testui.Ui{}
		strategy = cmd.NewBasicLoginStrategy(sessionFactory, config, testUI)
	})

	Describe("Try", func() {
		var (
			initialSession *fakecmd.FakeSession
			updatedSession *fakecmd.FakeSession
			updatedConfig  *fakecmdconf.FakeConfig
			director       *fakedir.FakeDirector
		)

		BeforeEach(func() {
			initialSession = &fakecmd.FakeSession{}
			sessions[config] = initialSession

			initialSession.EnvironmentReturns("environment")

			updatedConfig = &fakecmdconf.FakeConfig{}
			config.SetCredentialsStub = func(environment string, creds cmdconf.Creds) cmdconf.Config {
				updatedConfig.CredentialsStub = func(t string) cmdconf.Creds {
					return map[string]cmdconf.Creds{environment: creds}[t]
				}
				return updatedConfig
			}

			updatedSession = &fakecmd.FakeSession{}
			sessions[updatedConfig] = updatedSession

			director = &fakedir.FakeDirector{}
			updatedSession.DirectorReturns(director, nil)
		})

		act := func() error { return strategy.Try() }

		itLogsInOrErrs := func(expectedEnvironment, expectedUsername, expectedPassword string) {
			Context("when credentials are correct", func() {
				BeforeEach(func() {
					director.IsAuthenticatedReturns(true, nil)
				})

				It("successfully logs in", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(testUI.Said).To(Equal([]string{"Using environment 'environment'", fmt.Sprintf("Logged in to '%s'", expectedEnvironment)}))
				})

				It("saves the config with new credentials", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(updatedConfig.SaveCallCount()).To(Equal(1))
					Expect(updatedConfig.Credentials(expectedEnvironment)).To(Equal(cmdconf.Creds{
						Client:       expectedUsername,
						ClientSecret: expectedPassword,
					}))
				})
			})

			Context("when cannot check credentials correctness", func() {
				BeforeEach(func() {
					director.IsAuthenticatedReturns(false, errors.New("fake-err"))
				})

				It("returns an error and does not save config", func() {
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(updatedConfig.SaveCallCount()).To(Equal(0))
				})
			})
		}

		itKeepsAsking := func(expectedEnvironment, expectedUsername, expectedPassword string) {
			Context("when credentials are not correct", func() {
				BeforeEach(func() {
					tries := []bool{false, false, true}

					director.IsAuthenticatedStub = func() (bool, error) {
						result := tries[0]
						tries = tries[1:]
						return result, nil
					}
				})

				It("keeps on asking for new username and password until success", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(testUI.Errors).To(Equal([]string{
						"Failed to login to 'environment'",
						"Failed to login to 'environment'",
					}))

					Expect(testUI.Said).To(Equal([]string{"Using environment 'environment'", "Logged in to 'environment'"}))
				})

				It("only saves config upon successful log in", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(updatedConfig.SaveCallCount()).To(Equal(1))
					Expect(updatedConfig.Credentials(expectedEnvironment)).To(Equal(cmdconf.Creds{
						Client:       expectedUsername,
						ClientSecret: expectedPassword,
					}))
				})
			})
		}

		itErrsWithoutAsking := func() {
			Context("when credentials are not correct", func() {
				BeforeEach(func() {
					director.IsAuthenticatedReturns(false, nil)
				})

				It("returns an error without asking for username or password", func() {
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Invalid credentials"))

					Expect(testUI.Errors).To(Equal([]string{"Failed to login to 'environment'"}))
				})

				It("does not save config with new credentials", func() {
					err := act()
					Expect(err).To(HaveOccurred())

					Expect(updatedConfig.SaveCallCount()).To(Equal(0))
				})
			})
		}

		Context("when no global flags or config values are set", func() {
			BeforeEach(func() {
				testUI.AskedText = []testui.Answer{
					{Text: "asked-username1"},
					{Text: "asked-username2"},
					{Text: "asked-username3"},
				}

				testUI.AskedPasswords = []testui.Answer{
					{Text: "asked-password1"},
					{Text: "asked-password2"},
					{Text: "asked-password3"},
				}
			})

			itLogsInOrErrs("environment", "asked-username1", "asked-password1")
			itKeepsAsking("environment", "asked-username3", "asked-password3")
		})

		Context("when global flags or config values are set", func() {
			BeforeEach(func() {
				initialSession.CredentialsStub = func() cmdconf.Creds {
					return cmdconf.Creds{
						Client:       "global-username",
						ClientSecret: "global-password",
					}
				}
			})

			itLogsInOrErrs("environment", "global-username", "global-password")
			itErrsWithoutAsking()
		})
	})
})
