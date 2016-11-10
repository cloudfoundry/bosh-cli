package cmd_test

import (
	"errors"
	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AttachDisk", func() {
	var (
		director   *fakedir.FakeDirector
		command    AttachDiskCmd
		deployment *fakedir.FakeDeployment
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		deployment = &fakedir.FakeDeployment{}

		command = NewAttachDiskCmd(director, deployment)
	})

	Describe("Run", func() {
		var (
			opts         AttachDiskOpts
			act          func() error
			instanceSlug boshdir.InstanceSlug
			diskCid      string
		)

		BeforeEach(func() {
			act = func() error {
				err := command.Run(opts)
				return err
			}

			instanceSlug = boshdir.NewInstanceSlug("instance-group-name", "1")
			diskCid = "some-disk-id"

			opts = AttachDiskOpts{
				Args: AttachDiskArgs{
					Slug:   instanceSlug,
					DiskId: diskCid,
				},
			}
		})

		It("Tells the director to attach a disk", func() {
			act()
			Expect(director.AttachDiskCallCount()).To(Equal(1))

			receivedDeployemnt, receivedInstanceSlug, recievedDiskCid := director.AttachDiskArgsForCall(0)

			Expect(receivedDeployemnt).To(Equal(deployment))
			Expect(receivedInstanceSlug).To(Equal(instanceSlug))
			Expect(recievedDiskCid).To(Equal(diskCid))
		})

		Context("attaching a disk returns an error", func() {

			BeforeEach(func() {
				director.AttachDiskReturns(errors.New("director returned an error attaching a disk"))
			})

			It("Should return an error if director attaching a disk fails", func() {
				err := act()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
