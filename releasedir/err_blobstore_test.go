package releasedir_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
)

var _ = Describe("ErrBlobstore", func() {
	Describe("methods", func() {
		It("returns error", func() {
			blobErr := errors.New("fake-err")
			blob := NewErrBlobstore(blobErr)

			_, err := blob.Get("", boshcrypto.NewDigest(boshcrypto.DigestAlgorithmSHA1, ""))
			Expect(err).To(Equal(blobErr))

			_, _, err = blob.Create("")
			Expect(err).To(Equal(blobErr))

			err = blob.CleanUp("")
			Expect(err).To(Equal(blobErr))

			err = blob.Delete("")
			Expect(err).To(Equal(blobErr))

			err = blob.Validate()
			Expect(err).To(Equal(blobErr))
		})
	})
})
