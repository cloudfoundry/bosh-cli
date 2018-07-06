package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("DeploymentsCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  DeploymentsCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewDeploymentsCmd(ui, director)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("lists deployments", func() {
			deployments := []boshdir.DeploymentResp{
				boshdir.DeploymentResp{
					Name: "dep1",

					Teams: []string{"team1", "team2"},

					Releases: []boshdir.DeploymentReleaseResp{
						boshdir.DeploymentReleaseResp{
							Name:    "rel1",
							Version: "rel1-ver",
						},
						boshdir.DeploymentReleaseResp{
							Name:    "rel2",
							Version: "rel2-ver",
						},
					},

					Stemcells: []boshdir.DeploymentStemcellResp{
						boshdir.DeploymentStemcellResp{
							Name:    "stemcell1",
							Version: "stemcell1-ver",
						},
						boshdir.DeploymentStemcellResp{
							Name:    "stemcell2",
							Version: "stemcell2-ver",
						},
					},
				},
			}

			director.ListDeploymentsReturns(deployments, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table).To(Equal(boshtbl.Table{
				Content: "deployments",

				Header: []boshtbl.Header{
					boshtbl.NewHeader("Name"),
					boshtbl.NewHeader("Release(s)"),
					boshtbl.NewHeader("Stemcell(s)"),
					boshtbl.NewHeader("Team(s)"),
				},

				SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("dep1"),
						boshtbl.NewValueStrings([]string{"rel1/rel1-ver", "rel2/rel2-ver"}),
						boshtbl.NewValueStrings([]string{"stemcell1/stemcell1-ver", "stemcell2/stemcell2-ver"}),
						boshtbl.NewValueStrings([]string{"team1", "team2"}),
					},
				},
			}))
		})

		It("returns error if deployments cannot be retrieved", func() {
			director.ListDeploymentsReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
