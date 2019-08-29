package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
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
				Args: AllOrInstanceGroupOrInstanceSlugArgs{
					Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0"),
				},
			}
		})

		act := func() error { return command.Run(opts) }

		It("starts deployment, pool or instances", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StartCallCount()).To(Equal(1))
			slug, _ := deployment.StartArgsForCall(0)
			Expect(slug).To(Equal(
				boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0")))
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

		It("can set canaries", func() {
			opts.Canaries = "100%"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StartCallCount()).To(Equal(1))

			_, opts := deployment.StartArgsForCall(0)
			Expect(opts.Canaries).To(Equal("100%"))
		})

		It("can set max_in_flight", func() {
			opts.MaxInFlight = "5"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StartCallCount()).To(Equal(1))

			_, opts := deployment.StartArgsForCall(0)
			Expect(opts.MaxInFlight).To(Equal("5"))
		})
		Context("coverge and no-converge flags", func() {
			It("can set converge", func() {
				opts.Converge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StartCallCount()).To(Equal(1))

				_, opts := deployment.StartArgsForCall(0)
				Expect(opts.Converge).To(BeTrue())
			})

			It("converge by default", func() {
				opts.Converge = false
				opts.NoConverge = false
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StartCallCount()).To(Equal(1))

				_, opts := deployment.StartArgsForCall(0)
				Expect(opts.Converge).To(BeTrue())
			})

			It("can set no-converge", func() {
				opts.NoConverge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StartCallCount()).To(Equal(1))

				_, opts := deployment.StartArgsForCall(0)
				Expect(opts.Converge).To(BeFalse())
			})

			It("rejects combining converge and no-converge", func() {
				opts.Converge = true
				opts.NoConverge = true
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set converge and no-converge"))
				Expect(deployment.StartCallCount()).To(Equal(0))
			})

			It("doesn't allow canaries flag when no-converge is specified", func() {
				opts.NoConverge = true
				opts.Canaries = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set canaries and no-converge"))
				Expect(deployment.StartCallCount()).To(Equal(0))
			})

			It("doesn't allow max-in-flight flag when no-converge is specified", func() {
				opts.NoConverge = true
				opts.MaxInFlight = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set max-in-flight and no-converge"))
				Expect(deployment.StartCallCount()).To(Equal(0))
			})

			Context("with invalid slugs for no-converge on a deployment", func() {

				BeforeEach(func() {
					opts = StartOpts{
						Args: AllOrInstanceGroupOrInstanceSlugArgs{
							Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("", ""),
						},
					}
				})
				It("errors", func() {
					opts.NoConverge = true
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are trying to run start with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance."))
					Expect(deployment.StartCallCount()).To(Equal(0))
				})
			})
		})
	})
})
