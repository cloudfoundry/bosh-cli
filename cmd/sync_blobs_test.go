package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("SyncBlobsCmd", func() {
	var (
		blobsDir     *fakereldir.FakeBlobsDir
		command      cmd.SyncBlobsCmd
		numOfWorkers int
	)

	BeforeEach(func() {
		numOfWorkers = 5
		blobsDir = &fakereldir.FakeBlobsDir{}
		command = cmd.NewSyncBlobsCmd(blobsDir, numOfWorkers)
	})

	Describe("Run", func() {
		act := func() error {
			return command.Run()
		}

		It("downloads all blobs", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.SyncBlobsCallCount()).To(Equal(1))
			Expect(blobsDir.SyncBlobsArgsForCall(0)).To(Equal(5))
		})

		It("returns error if download fails", func() {
			blobsDir.SyncBlobsReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
