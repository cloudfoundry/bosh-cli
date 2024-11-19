package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("AttachDisk", func() {
	var (
		command    cmd.AttachDiskCmd
		deployment *fakedir.FakeDeployment
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}

		command = cmd.NewAttachDiskCmd(deployment)
	})

	Describe("Run", func() {
		var (
			attachDiskOpts opts.AttachDiskOpts
			act            func() error
			instanceSlug   boshdir.InstanceSlug
			diskCid        string
			diskProperties string
		)

		BeforeEach(func() {
			act = func() error {
				err := command.Run(attachDiskOpts)
				return err
			}

			instanceSlug = boshdir.NewInstanceSlug("instance-group-name", "1")
			diskCid = "some-disk-id"
			diskProperties = "copy"

			attachDiskOpts = opts.AttachDiskOpts{
				Args: opts.AttachDiskArgs{
					Slug:    instanceSlug,
					DiskCID: diskCid,
				},
			}
			attachDiskOpts.DiskProperties = diskProperties
		})

		It("tells the director to attach a disk", func() {
			err := act()
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.AttachDiskCallCount()).To(Equal(1))

			receivedInstanceSlug, receivedDiskCid, receivedDiskProperties := deployment.AttachDiskArgsForCall(0)

			Expect(receivedInstanceSlug).To(Equal(instanceSlug))
			Expect(receivedDiskCid).To(Equal(diskCid))
			Expect(receivedDiskProperties).To(Equal("copy"))
		})

		Context("attaching a disk returns an error", func() {

			BeforeEach(func() {
				deployment.AttachDiskReturns(errors.New("director returned an error attaching a disk"))
			})

			It("Should return an error if director attaching a disk fails", func() {
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("director returned an error attaching a disk"))
			})
		})
	})
})
