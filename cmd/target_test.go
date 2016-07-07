package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	fakecmdconf "github.com/cloudfoundry/bosh-init/cmd/config/fakes"
	fakecmd "github.com/cloudfoundry/bosh-init/cmd/fakes"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

var _ = Describe("TargetCmd", func() {
	var (
		sessions map[*fakecmdconf.FakeConfig2]*fakecmd.FakeSession
		config   *fakecmdconf.FakeConfig2
		ui       *fakeui.FakeUI
		command  TargetCmd
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
				TargetURL:    "curr-target-url",
				TargetCACert: "curr-ca-cert",
			},
		}

		ui = &fakeui.FakeUI{}

		command = NewTargetCmd(sessionFactory, config, ui)
	})

	Describe("Run", func() {
		var (
			opts            TargetOpts
			updatedSession  *fakecmd.FakeSession
			updatedConfig   *fakecmdconf.FakeConfig2
			updatedDirector *fakedir.FakeDirector
		)

		BeforeEach(func() {
			opts = TargetOpts{}
		})

		act := func() error { return command.Run(opts) }

		Context("when URL / name args are given without CA cert", func() {
			BeforeEach(func() {
				opts.Args.URL = "target-url"
				opts.Args.Alias = "target-alias"

				updatedConfig = &fakecmdconf.FakeConfig2{
					Existing: fakecmdconf.ConfigContents{
						TargetURL:   "target-url",
						TargetAlias: "target-alias",
					},
				}

				updatedDirector = &fakedir.FakeDirector{}

				updatedSession = &fakecmd.FakeSession{}
				updatedSession.DirectorReturns(updatedDirector, nil)
				updatedSession.TargetReturns("target-url")

				sessions[updatedConfig] = updatedSession
			})

			Context("when target is reachable", func() {
				BeforeEach(func() {
					info := boshdir.Info{
						Name:    "director-name",
						UUID:    "director-uuid",
						Version: "director-version",
					}
					updatedDirector.InfoReturns(info, nil)
				})

				It("sets and saves current target", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(config.Saved.TargetURL).To(Equal("target-url"))
					Expect(config.Saved.TargetAlias).To(Equal("target-alias"))
					Expect(config.Saved.TargetCACert).To(Equal(""))
				})

				It("shows current target and director info", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Said).To(Equal([]string{"Target set to 'target-url'"}))

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

			Context("when target is not reachable", func() {
				var (
					altUpdatedSession  *fakecmd.FakeSession
					altUpdatedConfig   *fakecmdconf.FakeConfig2
					altUpdatedDirector *fakedir.FakeDirector
				)

				BeforeEach(func() {
					updatedDirector.InfoReturns(boshdir.Info{}, errors.New("fake-err"))

					altUpdatedConfig = &fakecmdconf.FakeConfig2{
						Existing: fakecmdconf.ConfigContents{
							TargetURL:    "target-url",
							TargetAlias:  "target-alias",
							TargetCACert: "curr-ca-cert",
						},
					}

					altUpdatedDirector = &fakedir.FakeDirector{}

					altUpdatedSession = &fakecmd.FakeSession{}
					altUpdatedSession.DirectorReturns(altUpdatedDirector, nil)
					altUpdatedSession.TargetReturns("target-url")

					sessions[altUpdatedConfig] = altUpdatedSession
				})

				Context("when target using existing certificate is reachable", func() {
					BeforeEach(func() {
						info := boshdir.Info{
							Name:    "director-name",
							UUID:    "director-uuid",
							Version: "director-version",
						}
						altUpdatedDirector.InfoReturns(info, nil)
					})

					It("sets and saves current target with existing certificate", func() {
						err := act()
						Expect(err).ToNot(HaveOccurred())

						Expect(config.Saved.TargetURL).To(Equal("target-url"))
						Expect(config.Saved.TargetAlias).To(Equal("target-alias"))
						Expect(config.Saved.TargetCACert).To(Equal("curr-ca-cert"))
						Expect(config.Saved.Called).To(BeTrue())
					})

					It("shows current target and director info", func() {
						err := act()
						Expect(err).ToNot(HaveOccurred())

						Expect(ui.Said).To(Equal([]string{"Target set to 'target-url'"}))

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

				Context("when target using existing certificate is not reachable", func() {
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
				opts.Args.URL = "target-url"
				opts.Args.Alias = "target-alias"
				opts.CACert = FileBytesArg{Path: "target-ca-cert"}

				updatedConfig = &fakecmdconf.FakeConfig2{
					Existing: fakecmdconf.ConfigContents{
						TargetURL:    "target-url",
						TargetAlias:  "target-alias",
						TargetCACert: "target-ca-cert",
					},
				}

				updatedDirector = &fakedir.FakeDirector{}

				updatedSession = &fakecmd.FakeSession{}
				updatedSession.DirectorReturns(updatedDirector, nil)
				updatedSession.TargetReturns("target-url")

				sessions[updatedConfig] = updatedSession
			})

			Context("when target is reachable", func() {
				BeforeEach(func() {
					info := boshdir.Info{
						Name:    "director-name",
						UUID:    "director-uuid",
						Version: "director-version",
					}
					updatedDirector.InfoReturns(info, nil)
				})

				It("sets and saves current target", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(config.Saved.TargetURL).To(Equal("target-url"))
					Expect(config.Saved.TargetAlias).To(Equal("target-alias"))
					Expect(config.Saved.TargetCACert).To(Equal("target-ca-cert"))
				})

				It("shows current target and director info", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Said).To(Equal([]string{"Target set to 'target-url'"}))

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

			Context("when target is not reachable", func() {
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
				initialSession.TargetReturns("target-url")

				sessions[config] = initialSession
			})

			It("shows current target and director info", func() {
				info := boshdir.Info{
					Name:    "director-name",
					UUID:    "director-uuid",
					Version: "director-version",
				}

				director.InfoReturns(info, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Said).To(Equal([]string{"Current target is 'target-url'"}))

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
