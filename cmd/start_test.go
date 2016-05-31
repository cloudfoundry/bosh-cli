package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("StartCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    StartCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = NewStartCmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			opts StartOpts
		)

		BeforeEach(func() {
			opts = StartOpts{
				Args: AllOrPoolOrInstanceSlugArgs{
					Slug: boshdir.NewAllOrPoolOrInstanceSlug("some-name", ""),
				},
			}
		})

		act := func() error { return command.Run(opts) }

		It("starts deployment, pool or instances", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StartCallCount()).To(Equal(1))
			Expect(deployment.StartArgsForCall(0)).To(Equal(
				boshdir.NewAllOrPoolOrInstanceSlug("some-name", "")))
		})

		It("does not start if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.StartCallCount()).To(Equal(0))
		})

		It("returns error if start failed", func() {
			deployment.StartReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
