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

var _ = Describe("ProvideDiskCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.ProvideDiskCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewProvideDiskCmd(ui, director)
	})

	Describe("Run", func() {
		var provideDiskOpts opts.ProvideDiskOpts

		BeforeEach(func() {
		slug := boshdir.NewInstanceSlug("web", "abc123")

		provideDiskOpts = opts.ProvideDiskOpts{
			Args: opts.ProvideDiskArgs{
				DiskName:   "my-disk",
				InstanceID: slug,
			},
				DiskPool: "large",
				Size:     51200,
			}
		})

		act := func() error { return command.Run(provideDiskOpts) }

		It("provides (create+attach) a dynamic disk", func() {
			director.ProvideDynamicDiskReturns("disk-cid-123", nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			instanceID, diskName, diskPool, sizeInMB, metadata := director.ProvideDynamicDiskArgsForCall(0)
			Expect(instanceID).To(Equal("web/abc123"))
			Expect(diskName).To(Equal("my-disk"))
			Expect(diskPool).To(Equal("large"))
			Expect(sizeInMB).To(Equal(51200))
			Expect(metadata).To(BeNil())

			Expect(ui.Said).To(ContainElement(ContainSubstring("my-disk")))
		})

		It("returns error if providing disk failed", func() {
			director.ProvideDynamicDiskReturns("", errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
