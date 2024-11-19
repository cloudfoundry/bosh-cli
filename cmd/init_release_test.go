package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("InitReleaseCmd", func() {
	var (
		releaseDir *fakereldir.FakeReleaseDir
		command    cmd.InitReleaseCmd
	)

	BeforeEach(func() {
		releaseDir = &fakereldir.FakeReleaseDir{}
		command = cmd.NewInitReleaseCmd(releaseDir)
	})

	Describe("Run", func() {
		var (
			initReleaseOpts opts.InitReleaseOpts
		)

		BeforeEach(func() {
			initReleaseOpts = opts.InitReleaseOpts{}
		})

		act := func() error { return command.Run(initReleaseOpts) }

		It("inits release", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.InitCallCount()).To(Equal(1))
			Expect(releaseDir.InitArgsForCall(0)).To(BeFalse())
		})

		It("inits release with git as true", func() {
			initReleaseOpts.Git = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.InitCallCount()).To(Equal(1))
			Expect(releaseDir.InitArgsForCall(0)).To(BeTrue())
		})

		It("returns error if initing release fails", func() {
			releaseDir.InitReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
