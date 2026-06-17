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

var _ = Describe("CreateDynamicDiskCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.CreateDynamicDiskCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewCreateDynamicDiskCmd(ui, director)
	})

	Describe("Run", func() {
		var createOpts opts.CreateDynamicDiskOpts

		BeforeEach(func() {
			createOpts = opts.CreateDynamicDiskOpts{
				Args:     opts.CreateDynamicDiskArgs{DiskName: "my-disk"},
				DiskPool: "large",
				Size:     102400,
			}
		})

		act := func() error { return command.Run(createOpts) }

		It("creates a dynamic disk without attaching", func() {
			director.CreateDynamicDiskReturns("disk-cid-456", nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			diskName, diskPool, sizeInMB, metadata := director.CreateDynamicDiskArgsForCall(0)
			Expect(diskName).To(Equal("my-disk"))
			Expect(diskPool).To(Equal("large"))
			Expect(sizeInMB).To(Equal(102400))
			Expect(metadata).To(BeNil())

			Expect(ui.Said).To(ContainElement(ContainSubstring("my-disk")))
		})

		It("returns error if creation fails", func() {
			director.CreateDynamicDiskReturns("", errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
