package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/cmd/cmdfakes"
	fakecmdconf "github.com/cloudfoundry/bosh-cli/cmd/config/configfakes"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		session *fakecmd.FakeSession
		config  *fakecmdconf.FakeConfig
		ui      *fakeui.FakeUI
		command DeploymentCmd
	)

	BeforeEach(func() {
		session = &fakecmd.FakeSession{}
		config = &fakecmdconf.FakeConfig{}
		ui = &fakeui.FakeUI{}
		command = NewDeploymentCmd(session, config, ui)
	})

	Describe("Run", func() {
		var (
			deployment *fakedir.FakeDeployment
			director   *fakedir.FakeDirector
		)

		BeforeEach(func() {
			session.EnvironmentReturns("environment-url")
		})

		act := func() error { return command.Run() }

		Context("when director finds deployment", func() {
			It("shows current deployment name and list of configs", func() {
				deployment = &fakedir.FakeDeployment{
					NameStub: func() string { return "deployment-name" },
				}
				director = &fakedir.FakeDirector{}
				session.DeploymentReturns(deployment, nil)
				session.DirectorReturns(director, nil)
				director.ListDeploymentConfigsReturns(
					boshdir.DeploymentConfigs{
						Configs: []boshdir.DeploymentConfig{
							boshdir.DeploymentConfig{
								Config: boshdir.DeploymentConfigProperties{
									Id:   123,
									Type: "cloud",
									Name: "default",
								},
							}}}, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "deployments",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Release(s)"),
						boshtbl.NewHeader("Stemcell(s)"),
						boshtbl.NewHeader("Config(s)"),
						boshtbl.NewHeader("Team(s)"),
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("deployment-name"),
							boshtbl.NewValueStrings(nil),
							boshtbl.NewValueStrings(nil),
							boshtbl.NewValueStrings([]string{"123 cloud/default"}),
							boshtbl.NewValueStrings(nil),
						},
					},
				}))
			})
		})

		It("returns an error when director does not find deployment", func() {
			session.DeploymentReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(ui.Tables).To(BeEmpty())
		})
	})
})
