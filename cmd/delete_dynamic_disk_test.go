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

var _ = Describe("DeleteDynamicDiskCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.DeleteDynamicDiskCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewDeleteDynamicDiskCmd(ui, director)
	})

	Describe("Run", func() {
		var deleteOpts opts.DeleteDynamicDiskOpts

		BeforeEach(func() {
			deleteOpts = opts.DeleteDynamicDiskOpts{
				Args: opts.DeleteDynamicDiskArgs{DiskName: "my-disk"},
			}
		})

		act := func() error { return command.Run(deleteOpts) }

		It("deletes the dynamic disk", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.DeleteDynamicDiskArgsForCall(0)).To(Equal("my-disk"))
		})

		It("returns error if deletion fails", func() {
			director.DeleteDynamicDiskReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("does not delete if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.DeleteDynamicDiskCallCount()).To(Equal(0))
		})
	})
})
