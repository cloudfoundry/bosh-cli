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

var _ = Describe("StartCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    cmd.StartCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = cmd.NewStartCmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			startOpts opts.StartOpts
		)

		BeforeEach(func() {
			startOpts = opts.StartOpts{
				Args: opts.AllOrInstanceGroupOrInstanceSlugArgs{
					Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0"),
				},
			}
		})

		act := func() error { return command.Run(startOpts) }

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
			startOpts.Canaries = "100%"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StartCallCount()).To(Equal(1))

			_, startOpts := deployment.StartArgsForCall(0)
			Expect(startOpts.Canaries).To(Equal("100%"))
		})

		It("can set max_in_flight", func() {
			startOpts.MaxInFlight = "5"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.StartCallCount()).To(Equal(1))

			_, startOpts := deployment.StartArgsForCall(0)
			Expect(startOpts.MaxInFlight).To(Equal("5"))
		})
		Context("coverage and no-converge flags", func() {
			It("can set converge", func() {
				startOpts.Converge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StartCallCount()).To(Equal(1))

				_, startOpts := deployment.StartArgsForCall(0)
				Expect(startOpts.Converge).To(BeTrue())
			})

			It("converge by default", func() {
				startOpts.Converge = false
				startOpts.NoConverge = false
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StartCallCount()).To(Equal(1))

				_, startOpts := deployment.StartArgsForCall(0)
				Expect(startOpts.Converge).To(BeTrue())
			})

			It("can set no-converge", func() {
				startOpts.NoConverge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.StartCallCount()).To(Equal(1))

				_, startOpts := deployment.StartArgsForCall(0)
				Expect(startOpts.Converge).To(BeFalse())
			})

			It("rejects combining converge and no-converge", func() {
				startOpts.Converge = true
				startOpts.NoConverge = true
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set converge and no-converge"))
				Expect(deployment.StartCallCount()).To(Equal(0))
			})

			It("doesn't allow canaries flag when no-converge is specified", func() {
				startOpts.NoConverge = true
				startOpts.Canaries = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set canaries and no-converge"))
				Expect(deployment.StartCallCount()).To(Equal(0))
			})

			It("doesn't allow max-in-flight flag when no-converge is specified", func() {
				startOpts.NoConverge = true
				startOpts.MaxInFlight = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set max-in-flight and no-converge"))
				Expect(deployment.StartCallCount()).To(Equal(0))
			})

			Context("with invalid slugs for no-converge on a deployment", func() {

				BeforeEach(func() {
					startOpts = opts.StartOpts{
						Args: opts.AllOrInstanceGroupOrInstanceSlugArgs{
							Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("", ""),
						},
					}
				})
				It("errors", func() {
					startOpts.NoConverge = true
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are trying to run start with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance."))
					Expect(deployment.StartCallCount()).To(Equal(0))
				})
			})
		})
	})
})
