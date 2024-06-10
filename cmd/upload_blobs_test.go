package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("UploadBlobsCmd", func() {
	var (
		blobsDir *fakereldir.FakeBlobsDir
		command  cmd.UploadBlobsCmd
	)

	BeforeEach(func() {
		blobsDir = &fakereldir.FakeBlobsDir{}
		command = cmd.NewUploadBlobsCmd(blobsDir)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("uploads all blobs", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.UploadBlobsCallCount()).To(Equal(1))
		})

		It("returns error if upload fails", func() {
			blobsDir.UploadBlobsReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
