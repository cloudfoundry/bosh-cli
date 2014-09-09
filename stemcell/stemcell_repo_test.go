package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

var _ = Describe("Repo", func() {
	var (
		stemcellRepo   Repo
		fs             *fakesys.FakeFileSystem
		stemcellReader *fakebmstemcell.FakeStemcellReader
	)

	BeforeEach(func() {
		stemcellReader = fakebmstemcell.NewFakeReader()
		fs = fakesys.NewFakeFileSystem()
		stemcellRepo = NewRepo(fs, stemcellReader)
		fs.TempDirDir = "/path/to/dest"
	})

	It("Returns the extracted stemcell", func() {
		expectedStemcell := Stemcell{}
		stemcellReader.SetReadBehavior("/somepath", "/path/to/dest", expectedStemcell, nil)
		stemcell, extractedPath, err := stemcellRepo.Save("/somepath")

		Expect(err).ToNot(HaveOccurred())
		Expect(stemcell).To(Equal(expectedStemcell))
		Expect(extractedPath).To(Equal("/path/to/dest"))
	})

	It("return err when failed to create a tmpdir", func() {
		fs.TempDirError = errors.New("fake-fs-new-tempdir-error")
		_, _, err := stemcellRepo.Save("/somepath")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-fs-new-tempdir-error"))
	})

	It("return err and cleans up tmpdir when failed to read a stemcell tarball", func() {
		stemcellReader.SetReadBehavior("/somepath", "/path/to/dest", Stemcell{}, errors.New("fake-read-error"))
		_, _, err := stemcellRepo.Save("/somepath")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-read-error"))
		Expect(fs.FileExists("/path/to/dest")).To(BeFalse())
	})
})
