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

var _ = Describe("AttachDynamicDiskCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.AttachDynamicDiskCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewAttachDynamicDiskCmd(ui, director)
	})

	Describe("Run", func() {
		var attachOpts opts.AttachDynamicDiskOpts

		BeforeEach(func() {
		slug := boshdir.NewInstanceSlug("worker", "xyz789")

		attachOpts = opts.AttachDynamicDiskOpts{
			Args: opts.AttachDynamicDiskArgs{
				DiskName:   "my-disk",
				InstanceID: slug,
			},
			}
		})

		act := func() error { return command.Run(attachOpts) }

		It("attaches the dynamic disk to the instance", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			diskName, instanceID := director.AttachDynamicDiskArgsForCall(0)
			Expect(diskName).To(Equal("my-disk"))
			Expect(instanceID).To(Equal("worker/xyz789"))
		})

		It("returns error if attaching fails", func() {
			director.AttachDynamicDiskReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
