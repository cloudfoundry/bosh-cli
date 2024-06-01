package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("GeneratePackageCmd", func() {
	var (
		releaseDir *fakereldir.FakeReleaseDir
		command    cmd.GeneratePackageCmd
	)

	BeforeEach(func() {
		releaseDir = &fakereldir.FakeReleaseDir{}
		command = cmd.NewGeneratePackageCmd(releaseDir)
	})

	Describe("Run", func() {
		var (
			generatePackageOpts opts.GeneratePackageOpts
		)

		BeforeEach(func() {
			generatePackageOpts = opts.GeneratePackageOpts{Args: opts.GeneratePackageArgs{Name: "pkg"}}
		})

		act := func() error { return command.Run(generatePackageOpts) }

		It("generates package", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.GeneratePackageCallCount()).To(Equal(1))
			Expect(releaseDir.GeneratePackageArgsForCall(0)).To(Equal("pkg"))
		})

		It("returns error if generating package fails", func() {
			releaseDir.GeneratePackageReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
