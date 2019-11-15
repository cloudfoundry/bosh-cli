package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	realdirector "github.com/cloudfoundry/bosh-cli/director"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("CleanUpCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  CleanUpCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewCleanUpCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts CleanUpOpts
		)

		BeforeEach(func() {
			opts = CleanUpOpts{}
		})

		act := func() error { return command.Run(opts) }

		It("cleans up director resources", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.CleanUpCallCount()).To(Equal(1))
			Expect(director.CleanUpArgsForCall(0)).To(BeFalse())
		})

		It("cleans up *all* director resources", func() {
			opts.All = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.CleanUpCallCount()).To(Equal(1))
			Expect(director.CleanUpArgsForCall(0)).To(BeTrue())
		})

		It("does not clean up if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.CleanUpCallCount()).To(Equal(0))
		})

		It("returns error if cleaning up fails", func() {
			director.CleanUpReturns(realdirector.CleanUp{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("Print", func() {
		It("shows information about artifacts to be deleted", func() {
			cleanUp := realdirector.CleanUp{
				Releases: []string{"release1", "release2"},
			}

			command.PrintCleanUpTable(cleanUp)
			Expect(ui.Table.Header).To(Equal([]boshtbl.Header{
				boshtbl.NewHeader("Releases"),
				boshtbl.NewHeader("Stemcells"),
				boshtbl.NewHeader("Compiled Packages"),
				boshtbl.NewHeader("Orphaned Disks"),
				boshtbl.NewHeader("Orphaned VMs"),
				boshtbl.NewHeader("Exported Releases"),
				boshtbl.NewHeader("DNS Blobs"),
			}))
			Expect(ui.Table.Rows).To(HaveLen(1))

			Expect(ui.Table.Rows[0][0]).To(Equal(
				boshtbl.NewValueStrings(cleanUp.Releases),
			))
		})
	})
})
