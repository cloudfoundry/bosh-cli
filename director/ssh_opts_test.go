package director_test

import (
	"crypto/rsa"
	"errors"

	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"

	. "github.com/cloudfoundry/bosh-cli/director"
)

var _ = Describe("NewSSHOpts", func() {
	var (
		uuidGen *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		uuidGen = &fakeuuid.FakeGenerator{
			GeneratedUUID: "2a4e8104-dc50-4ad7-939a-2efd53b029ae",
		}
	})

	It("returns opts and private key", func() {
		opts, privKeyStr, err := NewSSHOpts(uuidGen)
		Expect(err).ToNot(HaveOccurred())

		Expect(opts.Username).To(Equal("bosh_2a4e8104dc504ad7"))
		Expect(opts.PublicKey).ToNot(BeEmpty())
		Expect(privKeyStr).ToNot(BeEmpty())

		privKey, err := ssh.ParseRawPrivateKey([]byte(privKeyStr))
		Expect(err).ToNot(HaveOccurred())
		Expect(privKey).ToNot(BeNil())

		pubKey, err := ssh.NewPublicKey(privKey.(*rsa.PrivateKey).Public())
		Expect(err).ToNot(HaveOccurred())

		Expect(opts.PublicKey).To(Equal(string(ssh.MarshalAuthorizedKey(pubKey))))
	})

	It("generates 2048 bits private key", func() {
		_, privKeyStr, err := NewSSHOpts(uuidGen)
		Expect(err).ToNot(HaveOccurred())

		privKey, err := ssh.ParseRawPrivateKey([]byte(privKeyStr))
		Expect(err).ToNot(HaveOccurred())

		Expect(privKey.(*rsa.PrivateKey).D.BitLen()).To(BeNumerically("~", 2048, 20))
	})

	It("returns error if uuid cannot be generated", func() {
		uuidGen.GenerateError = errors.New("fake-err")

		_, _, err := NewSSHOpts(uuidGen)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})
})
