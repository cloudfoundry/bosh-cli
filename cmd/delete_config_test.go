package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("DeleteConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  DeleteConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewDeleteConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts DeleteConfigOpts
		)

		BeforeEach(func() {
			opts = DeleteConfigOpts{
				Type: "my-type",
				Name: "my-name",
			}
		})

		act := func() error { return command.Run(opts) }

		It("does stop if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.DeleteConfigCallCount()).To(Equal(0))
		})

		Context("when neither ID nor options are given", func() {
			BeforeEach(func() {
				opts = DeleteConfigOpts{}
			})

			It("returns an error", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Either <ID> or parameters --type and --name must be provided"))
			})
		})

		Context("when only ID is given", func() {
			BeforeEach(func() {
				opts = DeleteConfigOpts{
					Args: DeleteConfigArgs{ID: "123"},
				}
			})

			Context("when there is a matching config", func() {
				It("succeeds", func() {
					director.DeleteConfigByIDReturns(true, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when there is NO matching config", func() {
				It("succeeds with a message as hint", func() {
					director.DeleteConfigByIDReturns(false, nil)

					err := act()
					Expect(err).To(Not(HaveOccurred()))
					Expect(ui.Said[0]).To(ContainSubstring("No configs to delete: no matches for id '123'"))
				})
			})
		})

		Context("when ID and type option is given", func() {

			BeforeEach(func() {
				opts = DeleteConfigOpts{
					Args: DeleteConfigArgs{ID: "123"},
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
				opts = DeleteConfigOpts{
					Args: DeleteConfigArgs{ID: "123"},
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
				opts = DeleteConfigOpts{
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
				opts = DeleteConfigOpts{
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

			Context("when there is a matching config", func() {
				It("succeeds", func() {
					director.DeleteConfigReturns(true, nil)

					err := act()
					Expect(err).To(Not(HaveOccurred()))
				})
			})

			Context("when there is NO matching config", func() {
				It("succeeds with a message as hint", func() {
					director.DeleteConfigReturns(false, nil)

					err := act()
					Expect(err).To(Not(HaveOccurred()))
					Expect(ui.Said[0]).To(ContainSubstring("No configs to delete: no matches for type 'my-type' and name 'my-name' found."))
				})
			})

			It("returns error if config cannot be deleted", func() {
				director.DeleteConfigReturns(false, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
