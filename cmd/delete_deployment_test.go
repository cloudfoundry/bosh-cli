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

var _ = Describe("DeleteDeploymentCmd", func() {
	var (
		testUI     *testui.Ui
		deployment *fakedir.FakeDeployment
		command    cmd.DeleteDeploymentCmd
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		deployment = &fakedir.FakeDeployment{}
		command = cmd.NewDeleteDeploymentCmd(testUI, deployment)
	})

	Describe("Run", func() {
		var (
			deleteDeploymentOpts opts.DeleteDeploymentOpts
		)

		BeforeEach(func() {
			deleteDeploymentOpts = opts.DeleteDeploymentOpts{}
		})

		act := func() error { return command.Run(deleteDeploymentOpts) }

		It("deletes deployment", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.DeleteCallCount()).To(Equal(1))
			Expect(deployment.DeleteArgsForCall(0)).To(BeFalse())
		})

		It("deletes deployment forcefully if requested", func() {
			deleteDeploymentOpts.Force = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.DeleteCallCount()).To(Equal(1))
			Expect(deployment.DeleteArgsForCall(0)).To(BeTrue())
		})

		It("does not delete deployment if confirmation is rejected", func() {
			testUI.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.DeleteCallCount()).To(Equal(0))
		})

		It("returns error if deleting deployment failed", func() {
			deployment.DeleteReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
