package tar_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/tar"
)

var _ = Describe("cmdExtractor", func() {
	var (
		fakeRunner *fakesys.FakeCmdRunner
		logger     boshlog.Logger
		extractor  Extractor
	)

	BeforeEach(func() {
		fakeRunner = fakesys.NewFakeCmdRunner()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		extractor = NewCmdExtractor(fakeRunner, logger)
	})

	Describe("Extract", func() {
		It("extracts the tar into the given dir", func() {
			err := extractor.Extract("/some/tar", "/some/extracted/dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeRunner.RunCommands).To(ContainElement([]string{"tar", "-C", "/some/extracted/dir", "-xzf", "/some/tar"}))
		})

		Context("when the tar cannot be extracted", func() {
			BeforeEach(func() {
				result := fakesys.FakeCmdResult{
					Error: errors.New(""),
				}
				fakeRunner.AddCmdResult("tar -C /some/extracted/dir -xzf /some/tar", result)
			})

			It("returns err", func() {
				err := extractor.Extract("/some/tar", "/some/extracted/dir")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Extracting tar `/some/tar' to `/some/extracted/dir'"))
			})
		})
	})
})
