package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("DeleteVMCmd", func() {
	var (
		testUI     *testui.Ui
		deployment *fakedir.FakeDeployment
		command    cmd.DeleteVMCmd
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		deployment = &fakedir.FakeDeployment{}
		command = cmd.NewDeleteVMCmd(testUI, deployment)
	})

	Describe("Run", func() {
		var (
			deleteVMOpts opts.DeleteVMOpts
		)

		BeforeEach(func() {
			deleteVMOpts = opts.DeleteVMOpts{
				Args: opts.DeleteVMArgs{CID: "some-cid"},
			}
		})

		act := func() error { return command.Run(deleteVMOpts) }

		It("deletes vm", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.DeleteVMCallCount()).To(Equal(1))
			Expect(deployment.DeleteVMArgsForCall(0)).To(Equal("some-cid"))
		})

		It("does not delete snapshot if confirmation is rejected", func() {
			testUI.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.DeleteVMCallCount()).To(Equal(0))
		})

		It("returns error if deleting snapshot failed", func() {
			deployment.DeleteVMReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
