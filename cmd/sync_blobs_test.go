package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	fakereldir "github.com/cloudfoundry/bosh-init/releasedir/fakes"
)

var _ = Describe("SyncBlobsCmd", func() {
	var (
		blobsDir *fakereldir.FakeBlobsDir
		command  SyncBlobsCmd
	)

	BeforeEach(func() {
		blobsDir = &fakereldir.FakeBlobsDir{}
		command = NewSyncBlobsCmd(blobsDir)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("downloads all blobs", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.DownloadBlobsCallCount()).To(Equal(1))
		})

		It("returns error if download fails", func() {
			blobsDir.DownloadBlobsReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
