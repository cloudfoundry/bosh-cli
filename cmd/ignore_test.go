package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("IgnoreCmd", func() {
	var (
		deployment *fakedir.FakeDeployment
		command    cmd.IgnoreCmd
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}
		command = cmd.NewIgnoreCmd(deployment)
	})

	Describe("Run", func() {
		var (
			ignoreOpts opts.IgnoreOpts
		)

		BeforeEach(func() {
			ignoreOpts = opts.IgnoreOpts{}
		})

		act := func() error {
			return command.Run(ignoreOpts)
		}

		Context("when ignoring an instance", func() {
			BeforeEach(func() {
				ignoreOpts.Args.Slug = boshdir.NewInstanceSlug("some-name", "some-id")
			})

			It("ignores the instance", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.IgnoreCallCount()).To(Equal(1))

				slugArg, ignoreArg := deployment.IgnoreArgsForCall(0)
				Expect(slugArg).To(Equal(boshdir.NewInstanceSlug("some-name", "some-id")))
				Expect(ignoreArg).To(Equal(true))
			})

			Context("when ignoring fails", func() {

				BeforeEach(func() {
					deployment.IgnoreReturns(errors.New("nope nope nope"))
				})

				It("returns the error", func() {
					err := act()
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
