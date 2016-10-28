package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("DeleteVmCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    DeleteVmCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = NewDeleteVmCmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			opts DeleteVmOpts
		)

		BeforeEach(func() {
			opts = DeleteVmOpts{
				Args: DeleteVmArgs{CID: "some-cid"},
			}
		})

		act := func() error { return command.Run(opts) }

		It("deletes vm", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.DeleteVmCallCount()).To(Equal(1))
			Expect(deployment.DeleteVmArgsForCall(0)).To(Equal("some-cid"))
		})

		It("does not delete snapshot if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.DeleteVmCallCount()).To(Equal(0))
		})

		It("returns error if deleting snapshot failed", func() {
			deployment.DeleteVmReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
