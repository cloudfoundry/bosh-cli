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

var _ = Describe("DetachDynamicDiskCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.DetachDynamicDiskCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewDetachDynamicDiskCmd(ui, director)
	})

	Describe("Run", func() {
		var detachOpts opts.DetachDynamicDiskOpts

		BeforeEach(func() {
			detachOpts = opts.DetachDynamicDiskOpts{
				Args: opts.DetachDynamicDiskArgs{DiskName: "my-disk"},
			}
		})

		act := func() error { return command.Run(detachOpts) }

		It("detaches the dynamic disk", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.DetachDynamicDiskArgsForCall(0)).To(Equal("my-disk"))
		})

		It("returns error if detaching fails", func() {
			director.DetachDynamicDiskReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("does not detach if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.DetachDynamicDiskCallCount()).To(Equal(0))
		})
	})
})
