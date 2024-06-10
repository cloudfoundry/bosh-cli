package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("GenerateJobCmd", func() {
	var (
		releaseDir *fakereldir.FakeReleaseDir
		command    cmd.GenerateJobCmd
	)

	BeforeEach(func() {
		releaseDir = &fakereldir.FakeReleaseDir{}
		command = cmd.NewGenerateJobCmd(releaseDir)
	})

	Describe("Run", func() {
		var (
			generateJobOpts opts.GenerateJobOpts
		)

		BeforeEach(func() {
			generateJobOpts = opts.GenerateJobOpts{Args: opts.GenerateJobArgs{Name: "job"}}
		})

		act := func() error { return command.Run(generateJobOpts) }

		It("generates job", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.GenerateJobCallCount()).To(Equal(1))
			Expect(releaseDir.GenerateJobArgsForCall(0)).To(Equal("job"))
		})

		It("returns error if generating job fails", func() {
			releaseDir.GenerateJobReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
