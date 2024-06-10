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

var _ = Describe("DeleteDiskCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.DeleteDiskCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewDeleteDiskCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			deleteDiskOpts opts.DeleteDiskOpts
		)

		BeforeEach(func() {
			deleteDiskOpts = opts.DeleteDiskOpts{
				Args: opts.DeleteDiskArgs{CID: "disk-cid"},
			}
		})

		act := func() error { return command.Run(deleteDiskOpts) }

		It("deletes orphaned disk", func() {
			disk := &fakedir.FakeOrphanDisk{}
			director.FindOrphanDiskReturns(disk, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.FindOrphanDiskArgsForCall(0)).To(Equal("disk-cid"))
			Expect(disk.DeleteCallCount()).To(Equal(1))
		})

		It("returns error if deleting disk failed", func() {
			disk := &fakedir.FakeOrphanDisk{}
			director.FindOrphanDiskReturns(disk, nil)

			disk.DeleteReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("does not delete disk if confirmation is rejected", func() {
			disk := &fakedir.FakeOrphanDisk{}
			director.FindOrphanDiskReturns(disk, nil)

			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(disk.DeleteCallCount()).To(Equal(0))
		})

		It("returns error if finding disk failed", func() {
			director.FindOrphanDiskReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
