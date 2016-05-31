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

var _ = Describe("RestartCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    RestartCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = NewRestartCmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			opts RestartOpts
		)

		BeforeEach(func() {
			opts = RestartOpts{
				Args: AllOrPoolOrInstanceSlugArgs{
					Slug: boshdir.NewAllOrPoolOrInstanceSlug("some-name", ""),
				},
			}
		})

		act := func() error { return command.Run(opts) }

		It("restarts deployment, pool or instances", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RestartCallCount()).To(Equal(1))

			slug, sd, force := deployment.RestartArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrPoolOrInstanceSlug("some-name", "")))
			Expect(sd).To(Equal(boshdir.SkipDrain{}))
			Expect(force).To(BeFalse())
		})

		It("restarts allowing to skip drain scripts", func() {
			opts.SkipDrain = boshdir.SkipDrain{All: true}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RestartCallCount()).To(Equal(1))

			slug, sd, force := deployment.RestartArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrPoolOrInstanceSlug("some-name", "")))
			Expect(sd).To(Equal(boshdir.SkipDrain{All: true}))
			Expect(force).To(BeFalse())
		})

		It("restarts forcefully", func() {
			opts.Force = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RestartCallCount()).To(Equal(1))

			slug, sd, force := deployment.RestartArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrPoolOrInstanceSlug("some-name", "")))
			Expect(sd).To(Equal(boshdir.SkipDrain{}))
			Expect(force).To(BeTrue())
		})

		It("does not restart if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.RestartCallCount()).To(Equal(0))
		})

		It("returns error if restart failed", func() {
			deployment.RestartReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
