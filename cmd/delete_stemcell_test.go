package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DeleteStemcellCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.DeleteStemcellCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewDeleteStemcellCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			deleteStemcellOpts opts.DeleteStemcellOpts
			stemcell           *fakedir.FakeStemcell
		)

		BeforeEach(func() {
			deleteStemcellOpts = opts.DeleteStemcellOpts{
				Args: opts.DeleteStemcellArgs{
					Slug: boshdir.NewStemcellSlug("some-name", "some-version"),
				},
			}

			stemcell = &fakedir.FakeStemcell{}
			director.FindStemcellReturns(stemcell, nil)
		})

		act := func() error { return command.Run(deleteStemcellOpts) }

		It("deletes stemcell", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.FindStemcellCallCount()).To(Equal(1))
			Expect(director.FindStemcellArgsForCall(0)).To(Equal(
				boshdir.NewStemcellSlug("some-name", "some-version")))

			Expect(stemcell.DeleteCallCount()).To(Equal(1))
			Expect(stemcell.DeleteArgsForCall(0)).To(BeFalse())
		})

		It("deletes stemcell forcefully if requested", func() {
			deleteStemcellOpts.Force = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(stemcell.DeleteCallCount()).To(Equal(1))
			Expect(stemcell.DeleteArgsForCall(0)).To(BeTrue())
		})

		It("does not delete stemcell if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(stemcell.DeleteCallCount()).To(Equal(0))
		})

		It("returns error if deleting stemcell failed", func() {
			stemcell.DeleteReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
