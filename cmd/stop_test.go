package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("StopCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    StopCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = NewStopCmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			opts StopOpts
		)

		BeforeEach(func() {
			opts = StopOpts{
				Args: AllOrInstanceGroupOrInstanceSlugArgs{
					Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0"),
				},
			}
		})

		act := func() error { return command.Run(opts) }

		It("stops deployment, pool or instances", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StopCallCount()).To(Equal(1))

			slug, stopOpts := deployment.StopArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0")))
			Expect(stopOpts.Hard).To(BeFalse())
			Expect(stopOpts.SkipDrain).To(BeFalse())
		})

		It("stops allowing to detach vms", func() {
			opts.Hard = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StopCallCount()).To(Equal(1))

			slug, stopOpts := deployment.StopArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0")))
			Expect(stopOpts.Hard).To(BeTrue())
			Expect(stopOpts.SkipDrain).To(BeFalse())
		})

		It("stops allowing to skip drain scripts", func() {
			opts.SkipDrain = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StopCallCount()).To(Equal(1))

			slug, stopOpts := deployment.StopArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0")))
			Expect(stopOpts.Hard).To(BeFalse())
			Expect(stopOpts.SkipDrain).To(BeTrue())
		})

		It("does not stop if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.StopCallCount()).To(Equal(0))
		})

		It("returns error if stop failed", func() {
			deployment.StopReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("can set canaries", func() {
			opts.Canaries = "30%"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StopCallCount()).To(Equal(1))

			_, stopOpts := deployment.StopArgsForCall(0)
			Expect(stopOpts.Canaries).To(Equal("30%"))
		})

		It("can set max_in_flight", func() {
			opts.MaxInFlight = "5"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StopCallCount()).To(Equal(1))

			_, stopOpts := deployment.StopArgsForCall(0)
			Expect(stopOpts.MaxInFlight).To(Equal("5"))
		})

		Context("coverge and no-converge flags", func() {
			It("can set converge", func() {
				opts.Converge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StopCallCount()).To(Equal(1))

				_, opts := deployment.StopArgsForCall(0)
				Expect(opts.Converge).To(BeTrue())
			})

			It("converge by default", func() {
				opts.Converge = false
				opts.NoConverge = false
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StopCallCount()).To(Equal(1))

				_, opts := deployment.StopArgsForCall(0)
				Expect(opts.Converge).To(BeTrue())
			})

			It("can set no-converge", func() {
				opts.NoConverge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StopCallCount()).To(Equal(1))

				_, opts := deployment.StopArgsForCall(0)
				Expect(opts.Converge).To(BeFalse())
			})

			It("rejects combining converge and no-converge", func() {
				opts.Converge = true
				opts.NoConverge = true
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set converge and no-converge"))
				Expect(deployment.StopCallCount()).To(Equal(0))
			})

			It("doesn't allow canaries flag when no-converge is specified", func() {
				opts.NoConverge = true
				opts.Canaries = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set canaries and no-converge"))
				Expect(deployment.StopCallCount()).To(Equal(0))
			})

			It("doesn't allow max-in-flight flag when no-converge is specified", func() {
				opts.NoConverge = true
				opts.MaxInFlight = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set max-in-flight and no-converge"))
				Expect(deployment.StopCallCount()).To(Equal(0))
			})

			Context("with invalid slugs for no-converge on a deployment", func() {

				BeforeEach(func() {
					opts = StopOpts{
						Args: AllOrInstanceGroupOrInstanceSlugArgs{
							Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("", ""),
						},
					}
				})
				It("errors", func() {
					opts.NoConverge = true
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are trying to run stop with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance."))
					Expect(deployment.StopCallCount()).To(Equal(0))
				})
			})
		})
	})
})
