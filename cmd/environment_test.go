package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	fakecmdconf "github.com/cloudfoundry/bosh-cli/cmd/config/fakes"
	fakecmd "github.com/cloudfoundry/bosh-cli/cmd/fakes"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("EnvironmentCmd", func() {
	var (
		sessions map[*fakecmdconf.FakeConfig2]*fakecmd.FakeSession
		config   *fakecmdconf.FakeConfig2
		ui       *fakeui.FakeUI
		command  EnvironmentCmd
	)

	BeforeEach(func() {
		sessions = map[*fakecmdconf.FakeConfig2]*fakecmd.FakeSession{}

		sessionFactory := func(config cmdconf.Config) Session {
			typedConfig, ok := config.(*fakecmdconf.FakeConfig2)
			if !ok {
				panic("Expected to find FakeConfig2")
			}

			for c, sess := range sessions {
				if c.Existing == typedConfig.Existing {
					return sess
				}
			}

			panic("Expected to find fake session")
		}

		config = &fakecmdconf.FakeConfig2{
			Existing: fakecmdconf.ConfigContents{
				EnvironmentURL:    "curr-environment-url",
				EnvironmentCACert: "curr-ca-cert",
			},
		}

		ui = &fakeui.FakeUI{}

		command = NewEnvironmentCmd(sessionFactory, config, ui)
	})

	Describe("Run", func() {
		var (
			opts            EnvironmentOpts
			updatedSession  *fakecmd.FakeSession
			updatedConfig   *fakecmdconf.FakeConfig2
			updatedDirector *fakedir.FakeDirector
		)

		BeforeEach(func() {
			opts = EnvironmentOpts{}
		})

		act := func() error { return command.Run(opts) }

		Context("when URL / name args are given without CA cert", func() {
			BeforeEach(func() {
				opts.Args.URL = "environment-url"
				opts.Args.Alias = "environment-alias"

				updatedConfig = &fakecmdconf.FakeConfig2{
					Existing: fakecmdconf.ConfigContents{
						EnvironmentURL:   "environment-url",
						EnvironmentAlias: "environment-alias",
					},
				}

				updatedDirector = &fakedir.FakeDirector{}

				updatedSession = &fakecmd.FakeSession{}
				updatedSession.DirectorReturns(updatedDirector, nil)
				updatedSession.EnvironmentReturns("environment-url")

				sessions[updatedConfig] = updatedSession
			})

			Context("when environment is reachable", func() {
				BeforeEach(func() {
					info := boshdir.Info{
						Name:    "director-name",
						UUID:    "director-uuid",
						Version: "director-version",
					}
					updatedDirector.InfoReturns(info, nil)
				})

				It("sets and saves current environment", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(config.Saved.EnvironmentURL).To(Equal("environment-url"))
					Expect(config.Saved.EnvironmentAlias).To(Equal("environment-alias"))
					Expect(config.Saved.EnvironmentCACert).To(Equal(""))
				})

				It("shows current environment and director info", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Said).To(Equal([]string{"Environment set to 'environment-url'"}))

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("Name"),
								boshtbl.NewValueString("director-name"),
							},
							{
								boshtbl.NewValueString("UUID"),
								boshtbl.NewValueString("director-uuid"),
							},
							{
								boshtbl.NewValueString("Version"),
								boshtbl.NewValueString("director-version"),
							},
							{
								boshtbl.NewValueString("User"),
								boshtbl.NewValueString("(not logged in)"),
							},
						},
					}))
				})
			})

			Context("when environment is not reachable", func() {
				var (
					altUpdatedSession  *fakecmd.FakeSession
					altUpdatedConfig   *fakecmdconf.FakeConfig2
					altUpdatedDirector *fakedir.FakeDirector
				)

				BeforeEach(func() {
					updatedDirector.InfoReturns(boshdir.Info{}, errors.New("fake-err"))

					altUpdatedConfig = &fakecmdconf.FakeConfig2{
						Existing: fakecmdconf.ConfigContents{
							EnvironmentURL:    "environment-url",
							EnvironmentAlias:  "environment-alias",
							EnvironmentCACert: "curr-ca-cert",
						},
					}

					altUpdatedDirector = &fakedir.FakeDirector{}

					altUpdatedSession = &fakecmd.FakeSession{}
					altUpdatedSession.DirectorReturns(altUpdatedDirector, nil)
					altUpdatedSession.EnvironmentReturns("environment-url")

					sessions[altUpdatedConfig] = altUpdatedSession
				})

				Context("when environment using existing certificate is reachable", func() {
					BeforeEach(func() {
						info := boshdir.Info{
							Name:    "director-name",
							UUID:    "director-uuid",
							Version: "director-version",
						}
						altUpdatedDirector.InfoReturns(info, nil)
					})

					It("sets and saves current environment with existing certificate", func() {
						err := act()
						Expect(err).ToNot(HaveOccurred())

						Expect(config.Saved.EnvironmentURL).To(Equal("environment-url"))
						Expect(config.Saved.EnvironmentAlias).To(Equal("environment-alias"))
						Expect(config.Saved.EnvironmentCACert).To(Equal("curr-ca-cert"))
						Expect(config.Saved.Called).To(BeTrue())
					})

					It("shows current environment and director info", func() {
						err := act()
						Expect(err).ToNot(HaveOccurred())

						Expect(ui.Said).To(Equal([]string{"Environment set to 'environment-url'"}))

						Expect(ui.Table).To(Equal(boshtbl.Table{
							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueString("Name"),
									boshtbl.NewValueString("director-name"),
								},
								{
									boshtbl.NewValueString("UUID"),
									boshtbl.NewValueString("director-uuid"),
								},
								{
									boshtbl.NewValueString("Version"),
									boshtbl.NewValueString("director-version"),
								},
								{
									boshtbl.NewValueString("User"),
									boshtbl.NewValueString("(not logged in)"),
								},
							},
						}))
					})
				})

				Context("when environment using existing certificate is not reachable", func() {
					BeforeEach(func() {
						altUpdatedDirector.InfoReturns(boshdir.Info{}, errors.New("fake-alt-err"))
					})

					It("returns an original error and does not save config", func() {
						err := act()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))

						Expect(config.Saved.Called).To(BeFalse())
					})
				})
			})
		})

		Context("when URL / name / CA cert args are given", func() {
			BeforeEach(func() {
				opts.Args.URL = "environment-url"
				opts.Args.Alias = "environment-alias"
				opts.CACert = "environment-ca-cert"

				updatedConfig = &fakecmdconf.FakeConfig2{
					Existing: fakecmdconf.ConfigContents{
						EnvironmentURL:    "environment-url",
						EnvironmentAlias:  "environment-alias",
						EnvironmentCACert: "environment-ca-cert",
					},
				}

				updatedDirector = &fakedir.FakeDirector{}

				updatedSession = &fakecmd.FakeSession{}
				updatedSession.DirectorReturns(updatedDirector, nil)
				updatedSession.EnvironmentReturns("environment-url")

				sessions[updatedConfig] = updatedSession
			})

			Context("when environment is reachable", func() {
				BeforeEach(func() {
					info := boshdir.Info{
						Name:    "director-name",
						UUID:    "director-uuid",
						Version: "director-version",
					}
					updatedDirector.InfoReturns(info, nil)
				})

				It("sets and saves environment environment", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(config.Saved.EnvironmentURL).To(Equal("environment-url"))
					Expect(config.Saved.EnvironmentAlias).To(Equal("environment-alias"))
					Expect(config.Saved.EnvironmentCACert).To(Equal("environment-ca-cert"))
				})

				It("shows current environment and director info", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Said).To(Equal([]string{"Environment set to 'environment-url'"}))

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("Name"),
								boshtbl.NewValueString("director-name"),
							},
							{
								boshtbl.NewValueString("UUID"),
								boshtbl.NewValueString("director-uuid"),
							},
							{
								boshtbl.NewValueString("Version"),
								boshtbl.NewValueString("director-version"),
							},
							{
								boshtbl.NewValueString("User"),
								boshtbl.NewValueString("(not logged in)"),
							},
						},
					}))
				})
			})

			Context("when environment is not reachable", func() {
				BeforeEach(func() {
					updatedDirector.InfoReturns(boshdir.Info{}, errors.New("fake-err"))
				})

				It("returns an error and does not save config", func() {
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(config.Saved.Called).To(BeFalse())
				})
			})
		})

		Context("when no args are given", func() {
			var (
				director *fakedir.FakeDirector
			)

			BeforeEach(func() {
				director = &fakedir.FakeDirector{}

				initialSession := &fakecmd.FakeSession{}
				initialSession.DirectorReturns(director, nil)
				initialSession.EnvironmentReturns("environment-url")

				sessions[config] = initialSession
			})

			It("shows current environment and director info", func() {
				info := boshdir.Info{
					Name:    "director-name",
					UUID:    "director-uuid",
					Version: "director-version",
				}

				director.InfoReturns(info, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Said).To(Equal([]string{"Current environment is 'environment-url'"}))

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("Name"),
							boshtbl.NewValueString("director-name"),
						},
						{
							boshtbl.NewValueString("UUID"),
							boshtbl.NewValueString("director-uuid"),
						},
						{
							boshtbl.NewValueString("Version"),
							boshtbl.NewValueString("director-version"),
						},
						{
							boshtbl.NewValueString("User"),
							boshtbl.NewValueString("(not logged in)"),
						},
					},
				}))
			})

			It("does not save config", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(config.Saved).To(BeNil())
			})

			It("returns error when cannot fetch director info", func() {
				director.InfoReturns(boshdir.Info{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
