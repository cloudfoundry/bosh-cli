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

var _ = Describe("EnvironmentCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  EnvironmentCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewEnvironmentCmd(ui, director)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("shows director info if able to fetch it", func() {
			info := boshdir.Info{
				Name:    "director-name",
				UUID:    "director-uuid",
				Version: "director-version",
			}
			director.InfoReturns(info, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

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

		It("returns error if director info cannot be fetched", func() {
			director.InfoReturns(boshdir.Info{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
