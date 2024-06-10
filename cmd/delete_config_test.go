package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DeleteConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.DeleteConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewDeleteConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			deleteConfigOpts opts.DeleteConfigOpts
		)

		BeforeEach(func() {
			deleteConfigOpts = opts.DeleteConfigOpts{
				Type: "my-type",
				Name: "my-name",
			}
		})

		act := func() error { return command.Run(deleteConfigOpts) }

		It("does stop if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.DeleteConfigCallCount()).To(Equal(0))
		})

		Context("when neither ID nor options are given", func() {
			BeforeEach(func() {
				deleteConfigOpts = opts.DeleteConfigOpts{}
			})

			It("returns an error", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Either <ID> or parameters --type and --name must be provided"))
			})
		})

		Context("when only ID is given", func() {
			BeforeEach(func() {
				deleteConfigOpts = opts.DeleteConfigOpts{
					Args: opts.DeleteConfigArgs{ID: "123"},
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
				deleteConfigOpts = opts.DeleteConfigOpts{
					Args: opts.DeleteConfigArgs{ID: "123"},
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
				deleteConfigOpts = opts.DeleteConfigOpts{
					Args: opts.DeleteConfigArgs{ID: "123"},
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
				deleteConfigOpts = opts.DeleteConfigOpts{
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
				deleteConfigOpts = opts.DeleteConfigOpts{
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
