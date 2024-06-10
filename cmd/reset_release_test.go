package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("ResetReleaseCmd", func() {
	var (
		releaseDir *fakereldir.FakeReleaseDir
		command    cmd.ResetReleaseCmd
	)

	BeforeEach(func() {
		releaseDir = &fakereldir.FakeReleaseDir{}
		command = cmd.NewResetReleaseCmd(releaseDir)
	})

	Describe("Run", func() {
		var (
			resetReleaseOpts opts.ResetReleaseOpts
		)

		BeforeEach(func() {
			resetReleaseOpts = opts.ResetReleaseOpts{}
		})

		act := func() error { return command.Run(resetReleaseOpts) }

		It("resets release", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.ResetCallCount()).To(Equal(1))
		})

		It("returns error if resetting release fails", func() {
			releaseDir.ResetReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
