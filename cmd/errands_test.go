package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("ErrandsCmd", func() {
	var (
		testUI     *testui.Ui
		deployment *fakedir.FakeDeployment
		command    cmd.ErrandsCmd
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		deployment = &fakedir.FakeDeployment{}
		command = cmd.NewErrandsCmd(testUI, deployment)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("lists current errands", func() {
			errands := []boshdir.Errand{
				{
					Name: "some-errand",
				},
			}

			deployment.ErrandsReturns(errands, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Table).To(Equal(boshtbl.Table{
				Content: "errands",

				Header: []boshtbl.Header{boshtbl.NewHeader("Name")},

				SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

				Rows: [][]boshtbl.Value{
					{boshtbl.NewValueString("some-errand")},
				},
			}))
		})

		It("returns error if errands cannot be retrieved", func() {
			deployment.ErrandsReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
