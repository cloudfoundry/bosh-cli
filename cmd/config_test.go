package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("ConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  ConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts ConfigOpts
		)

		act := func() error { return command.Run(opts) }

		Context("when neither ID nor options are given", func() {

			BeforeEach(func() {
				opts = ConfigOpts{}
			})

			It("returns an error", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Either <ID> or parameters --type and --name must be provided"))
			})
		})

		Context("when only ID is given", func() {

			BeforeEach(func() {
				opts = ConfigOpts{
					Args: ConfigArgs{ID: "123"},
				}
			})

			It("shows config if ID is correct", func() {
				config := boshdir.Config{
					ID:      "123",
					Content: "some-content",
				}

				director.LatestConfigByIDReturns(config, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(ui.Table).To(Equal(
					boshtbl.Table{
						Content: "config",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("ID"),
							boshtbl.NewHeader("Type"),
							boshtbl.NewHeader("Name"),
							boshtbl.NewHeader("Created At"),
							boshtbl.NewHeader("Content"),
						},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("123"),
								boshtbl.NewValueString(""),
								boshtbl.NewValueString(""),
								boshtbl.NewValueString(""),
								boshtbl.NewValueString("some-content"),
							},
						},

						Notes: []string{},

						FillFirstColumn: true,

						Transpose: true,
					}))
			})

			It("returns error if config cannot be retrieved", func() {
				director.LatestConfigByIDReturns(boshdir.Config{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when ID and type option is given", func() {

			BeforeEach(func() {
				opts = ConfigOpts{
					Args: ConfigArgs{ID: "123"},
					Type: "my-type",
				}
			})

			It("returns an error", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can only specify one of ID or parameters '--type' and '--name'"))
			})
		})

		Context("when ID and name option is given", func() {
			BeforeEach(func() {
				opts = ConfigOpts{
					Args: ConfigArgs{ID: "123"},
					Name: "my-name",
				}
			})

			It("returns an error", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can only specify one of ID or parameters '--type' and '--name'"))
			})
		})

		Context("when only the name option is given", func() {
			BeforeEach(func() {
				opts = ConfigOpts{
					Name: "my-name",
				}
			})

			It("returns an error", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Need to specify both parameters '--type' and '--name'"))
			})
		})

		Context("when only the type option is given", func() {
			BeforeEach(func() {
				opts = ConfigOpts{
					Type: "my-type",
				}
			})

			It("returns an error", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Need to specify both parameters '--type' and '--name'"))
			})
		})

		Context("when ID is not given and both options are given", func() {

			BeforeEach(func() {
				opts = ConfigOpts{
					Type: "my-type",
					Name: "my-name",
				}
			})

			It("shows config", func() {
				config := boshdir.Config{
					ID:      "123",
					Type:    "my-type",
					Name:    "my-name",
					Content: "some-content",
				}

				director.LatestConfigReturns(config, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(ui.Table).To(Equal(
					boshtbl.Table{
						Content: "config",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("ID"),
							boshtbl.NewHeader("Type"),
							boshtbl.NewHeader("Name"),
							boshtbl.NewHeader("Created At"),
							boshtbl.NewHeader("Content"),
						},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("123"),
								boshtbl.NewValueString("my-type"),
								boshtbl.NewValueString("my-name"),
								boshtbl.NewValueString(""),
								boshtbl.NewValueString("some-content"),
							},
						},

						Notes: []string{},

						FillFirstColumn: true,

						Transpose: true,
					}))
			})

			It("returns error if config cannot be retrieved", func() {
				director.LatestConfigReturns(boshdir.Config{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

	})
})
