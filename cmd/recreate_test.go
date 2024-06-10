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

var _ = Describe("RecreateCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    cmd.RecreateCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = cmd.NewRecreateCmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			recreateOpts opts.RecreateOpts
		)

		BeforeEach(func() {
			recreateOpts = opts.RecreateOpts{
				Args: opts.AllOrInstanceGroupOrInstanceSlugArgs{
					Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0"),
				},
			}
		})

		act := func() error { return command.Run(recreateOpts) }

		It("recreate deployment, pool or instances", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RecreateCallCount()).To(Equal(1))

			slug, recreateOpts := deployment.RecreateArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0")))
			Expect(recreateOpts.SkipDrain).To(BeFalse())
			Expect(recreateOpts.Force).To(BeFalse())
		})

		It("recreate allowing to skip drain scripts", func() {
			recreateOpts.SkipDrain = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RecreateCallCount()).To(Equal(1))

			slug, recreateOpts := deployment.RecreateArgsForCall(0)
			Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("some-name", "0")))
			Expect(recreateOpts.SkipDrain).To(BeTrue())
			Expect(recreateOpts.Force).To(BeFalse())
		})

		It("can set canaries", func() {
			recreateOpts.Canaries = "3"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RecreateCallCount()).To(Equal(1))

			_, recreateOpts := deployment.RecreateArgsForCall(0)
			Expect(recreateOpts.Canaries).To(Equal("3"))
		})

		It("can set max_in_flight", func() {
			recreateOpts.MaxInFlight = "5"

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RecreateCallCount()).To(Equal(1))

			_, recreateOpts := deployment.RecreateArgsForCall(0)
			Expect(recreateOpts.MaxInFlight).To(Equal("5"))
		})

		It("can set dry_run", func() {
			recreateOpts.DryRun = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RecreateCallCount()).To(Equal(1))

			_, recreateOpts := deployment.RecreateArgsForCall(0)
			Expect(recreateOpts.DryRun).To(BeTrue())
		})

		It("can set fix", func() {
			recreateOpts.Fix = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RecreateCallCount()).To(Equal(1))

			_, recreateOpts := deployment.RecreateArgsForCall(0)
			Expect(recreateOpts.Fix).To(BeTrue())
		})

		It("does not recreate if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.RecreateCallCount()).To(Equal(0))
		})

		It("returns error if restart failed", func() {
			deployment.RecreateReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		Context("coverge and no-converge flags", func() {
			It("can set converge", func() {
				recreateOpts.Converge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RecreateCallCount()).To(Equal(1))

				_, recreateOpts := deployment.RecreateArgsForCall(0)
				Expect(recreateOpts.Converge).To(BeTrue())
			})

			It("converge by default", func() {
				recreateOpts.Converge = false
				recreateOpts.NoConverge = false
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RecreateCallCount()).To(Equal(1))

				_, recreateOpts := deployment.RecreateArgsForCall(0)
				Expect(recreateOpts.Converge).To(BeTrue())
			})

			It("can set no-converge", func() {
				recreateOpts.NoConverge = true
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RecreateCallCount()).To(Equal(1))

				_, recreateOpts := deployment.RecreateArgsForCall(0)
				Expect(recreateOpts.Converge).To(BeFalse())
			})

			It("rejects combining converge and no-converge", func() {
				recreateOpts.Converge = true
				recreateOpts.NoConverge = true
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set converge and no-converge"))
				Expect(deployment.RecreateCallCount()).To(Equal(0))
			})

			It("doesn't allow canaries flag when no-converge is specified", func() {
				recreateOpts.NoConverge = true
				recreateOpts.Canaries = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set canaries and no-converge"))
				Expect(deployment.RecreateCallCount()).To(Equal(0))
			})

			It("doesn't allow max-in-flight flag when no-converge is specified", func() {
				recreateOpts.NoConverge = true
				recreateOpts.MaxInFlight = "1"
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set max-in-flight and no-converge"))
				Expect(deployment.RecreateCallCount()).To(Equal(0))
			})

			It("doesn't allow dry-run flag when no-converge is specified", func() {
				recreateOpts.NoConverge = true
				recreateOpts.DryRun = true
				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Can't set dry-run and no-converge"))
				Expect(deployment.RecreateCallCount()).To(Equal(0))
			})

			Context("with invalid slugs for no-converge on a deployment", func() {

				BeforeEach(func() {
					recreateOpts = opts.RecreateOpts{
						Args: opts.AllOrInstanceGroupOrInstanceSlugArgs{
							Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("", ""),
						},
					}
				})
				It("errors", func() {
					recreateOpts.NoConverge = true
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("You are trying to run recreate with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance."))
					Expect(deployment.RecreateCallCount()).To(Equal(0))
				})
			})
		})
	})
})
