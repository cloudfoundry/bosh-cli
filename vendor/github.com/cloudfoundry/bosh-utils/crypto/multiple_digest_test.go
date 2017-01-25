package crypto_test

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-utils/crypto"
)

var _ = Describe("MultipleDigest", func() {
	var (
		digest MultipleDigest
	)

	BeforeEach(func() {
		digest = MultipleDigest{}
	})

	Describe("Verify", func() {
		Context("for a multi digest containing no digests", func() {
			It("returns error", func() {
				err := MultipleDigest{}.Verify(strings.NewReader("desired content"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected to find at least one digest"))
			})
		})

		Context("for a multi digest containing only SHA1 digest", func() {
			BeforeEach(func() {
				abcDigest, err := DigestAlgorithmSHA1.CreateDigest(strings.NewReader("desired content"))
				Expect(err).ToNot(HaveOccurred())
				digest = MustNewMultipleDigest(abcDigest)
			})

			It("does not error when the checksum matches", func() {
				Expect(digest.Verify(strings.NewReader("desired content"))).ToNot(HaveOccurred())
			})

			It("errors when the checksum does not match", func() {
				err := digest.Verify(strings.NewReader("non-matching content"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected stream to have digest 'ab78f75acac9f803cf5948e2bce4100734d08bc1' but was '78f4f37d56ce7bcdcda243b60a09310a174977e3'"))
			})
		})

		Context("for a multi digest containing many digests", func() {
			Context("when the strongest digest matches", func() {
				BeforeEach(func() {
					sha1DesiredContentDigest, err := DigestAlgorithmSHA1.CreateDigest(strings.NewReader("weak digest content"))
					Expect(err).ToNot(HaveOccurred())
					sha256DesiredContentDigest, err := DigestAlgorithmSHA256.CreateDigest(strings.NewReader("weak digest content"))
					Expect(err).ToNot(HaveOccurred())
					sha512DesiredContentDigest, err := DigestAlgorithmSHA512.CreateDigest(strings.NewReader("strong desired content"))
					Expect(err).ToNot(HaveOccurred())
					digest = MustNewMultipleDigest(sha1DesiredContentDigest, sha256DesiredContentDigest, sha512DesiredContentDigest)
				})

				It("uses the strongest digest and does not error", func() {
					Expect(digest.Verify(strings.NewReader("strong desired content"))).ToNot(HaveOccurred())
				})

				It("returns errors when the content does not match the strongest digest (even if it does match weaker digests)", func() {
					err := digest.Verify(strings.NewReader("weak digest content"))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Expected stream to have digest 'sha512:df1f95d9baa88052449120ada4a32aef23e8b69d6f96f888ec5f79da43916595a416d76fb1f7cf4b9696cefdf200f3c506228616eb7ba911a7dbc8b0b0763b9f' but was 'sha512:eb9cec8ded76063096b71782234c875880da15d751df46a17ceae9ac68cb264d1bec3e062150ec7c8ba71249052b0c4118f5e8b1fdce945bb180d65604774884'"))
				})
			})

			Context("algorithm precedence", func() {
				It("uses sha256 over sha1", func() {
					sha1DesiredContentDigest, err := DigestAlgorithmSHA1.CreateDigest(strings.NewReader("weak digest content"))
					Expect(err).ToNot(HaveOccurred())
					sha256DesiredContentDigest, err := DigestAlgorithmSHA256.CreateDigest(strings.NewReader("strong digest content"))
					Expect(err).ToNot(HaveOccurred())
					digest = MustNewMultipleDigest(sha1DesiredContentDigest, sha256DesiredContentDigest)

					Expect(digest.Verify(strings.NewReader("strong digest content"))).ToNot(HaveOccurred())
				})

				It("uses sha512 over sha256 and sha1", func() {
					sha1DesiredContentDigest, err := DigestAlgorithmSHA1.CreateDigest(strings.NewReader("weak digest content"))
					Expect(err).ToNot(HaveOccurred())
					sha256DesiredContentDigest, err := DigestAlgorithmSHA256.CreateDigest(strings.NewReader("weak digest content"))
					Expect(err).ToNot(HaveOccurred())
					sha512DesiredContentDigest, err := DigestAlgorithmSHA512.CreateDigest(strings.NewReader("strong digest content"))
					Expect(err).ToNot(HaveOccurred())
					digest = MustNewMultipleDigest(sha1DesiredContentDigest, sha256DesiredContentDigest, sha512DesiredContentDigest)

					Expect(digest.Verify(strings.NewReader("strong digest content"))).ToNot(HaveOccurred())
				})

				It("uses sha1 over unknown algos", func() {
					unknown1Digest := NewDigest(NewUnknownAlgorithm("unknown1"), "val1")
					unknown2Digest := NewDigest(NewUnknownAlgorithm("unknown2"), "val2")
					sha1DesiredContentDigest, err := DigestAlgorithmSHA1.CreateDigest(strings.NewReader("strong digest content"))
					Expect(err).ToNot(HaveOccurred())
					digest = MustNewMultipleDigest(unknown1Digest, sha1DesiredContentDigest, unknown2Digest)

					Expect(digest.Verify(strings.NewReader("strong digest content"))).ToNot(HaveOccurred())
				})
			})

			It("returns an error if none of the algos are known", func() {
				unknown1Digest := NewDigest(NewUnknownAlgorithm("unknown1"), "val1")
				unknown2Digest := NewDigest(NewUnknownAlgorithm("unknown2"), "val2")
				digest = MustNewMultipleDigest(unknown1Digest, unknown2Digest)

				err := digest.Verify(strings.NewReader("strong digest content"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Computing digest from stream: Unable to create digest of unkown algorithm 'unknown1'"))
			})

			Context("when two of the digests are the same algorithm", func() {
				BeforeEach(func() {
					sha1DesiredContentDigestA, err := DigestAlgorithmSHA1.CreateDigest(strings.NewReader("digest content A"))
					Expect(err).ToNot(HaveOccurred())
					sha1DesiredContentDigestB, err := DigestAlgorithmSHA1.CreateDigest(strings.NewReader("digest content B"))
					Expect(err).ToNot(HaveOccurred())
					digest = MustNewMultipleDigest(sha1DesiredContentDigestA, sha1DesiredContentDigestB)
				})

				It("returns error stating that same algo types were not expected", func() {
					err := digest.Verify(strings.NewReader("digest content A"))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Multiple digests of the same algorithm 'sha1' found in digests 'cf305610f87bdfb86b0cf6aa01abeeed7411d1cc;e136b264965d153f51136924a93a855b2e841139'"))

					err = digest.Verify(strings.NewReader("digest content B"))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Multiple digests of the same algorithm 'sha1' found in digests 'cf305610f87bdfb86b0cf6aa01abeeed7411d1cc;e136b264965d153f51136924a93a855b2e841139'"))
				})
			})
		})
	})

	Describe("FullString", func() {
		It("returns the digest matching the algorithm", func() {
			digest1 := NewDigest(DigestAlgorithmSHA1, "sha1digestval")
			digest2 := NewDigest(DigestAlgorithmSHA256, "sha256digestval")
			digest := MustNewMultipleDigest(digest1, digest2)

			fullString := digest.FullString()
			Expect(fullString).To(Equal("sha1digestval;sha256:sha256digestval"))
		})
	})

	Describe("DigestFor", func() {
		Context("when the algorithm matches one of the digests in the multi", func () {
			It("returns the digest matching the algorithm", func() {
				digest1 := NewDigest(DigestAlgorithmSHA1, "sha1digestval")
				digest2 := NewDigest(DigestAlgorithmSHA256, "sha256digestval")
				digests := MustNewMultipleDigest(digest1, digest2)

				digest, err := digests.DigestFor(DigestAlgorithmSHA1)
				Expect(err).ToNot(HaveOccurred())
				Expect(digest).To(Equal(digest1))

				digest, err = digests.DigestFor(DigestAlgorithmSHA256)
				Expect(err).ToNot(HaveOccurred())
				Expect(digest).To(Equal(digest2))
			})
		})

		Context("when the algorithm specified does not match any contained digests", func () {
			It("returns an error", func () {
				digest1 := NewDigest(DigestAlgorithmSHA1, "sha1digestval")
				digest2 := NewDigest(DigestAlgorithmSHA256, "sha256digestval")
				digests := MustNewMultipleDigest(digest1, digest2)

				_, err := digests.DigestFor(DigestAlgorithmSHA512)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("MarshalJSON", func() {
		It("returns semicolon separated strings", func() {
			digest1 := NewDigest(DigestAlgorithmSHA1, "sha1digestval")
			digest2 := NewDigest(DigestAlgorithmSHA256, "sha256digestval")
			digest := MustNewMultipleDigest(digest1, digest2)

			bytes, err := json.Marshal(digest)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(Equal(`"sha1digestval;sha256:sha256digestval"`))
		})

		It("does not include sha1 prefix", func() {
			bytes, err := json.Marshal(MustNewMultipleDigest(NewDigest(DigestAlgorithmSHA1, "digestval")))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(Equal(`"digestval"`))
		})

		It("includes non-sha1 algo prefixes", func() {
			bytes, err := json.Marshal(MustNewMultipleDigest(NewDigest(DigestAlgorithmSHA256, "digestval")))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(Equal(`"sha256:digestval"`))
		})

		It("maintains order of digests", func() {
			digest1 := NewDigest(DigestAlgorithmSHA1, "digestval")
			digest2 := NewDigest(DigestAlgorithmSHA256, "digestval256")

			bytes, err := MustNewMultipleDigest(digest1, digest2).MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(Equal(`"digestval;sha256:digestval256"`))

			bytes, err = MustNewMultipleDigest(digest2, digest1).MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(Equal(`"sha256:digestval256;digestval"`))
		})

		It("retains unknown algos", func() {
			err := json.Unmarshal([]byte(`"unknown1:val1;sha256:val256"`), &digest)
			Expect(err).ToNot(HaveOccurred())

			bytes, err := json.Marshal(digest)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(Equal(`"unknown1:val1;sha256:val256"`))
		})
	})

	Describe("UnmarshalJSON", func() {
		It("parses from valid JSON picking strongest digest", func() {
			err := json.Unmarshal([]byte(`"sha1:abcdefg;sha256:1bf4b70c96b9d4e8f473ac6b7e6b5b965ab3497287a86eb2ed1b263287c78038"`), &digest)
			Expect(err).ToNot(HaveOccurred())

			Expect(digest.Algorithm().Name()).To(Equal(DigestAlgorithmSHA256.Name()))
			Expect(digest.String()).To(Equal("sha256:1bf4b70c96b9d4e8f473ac6b7e6b5b965ab3497287a86eb2ed1b263287c78038"))
			Expect(digest.Verify(strings.NewReader("content to be verified"))).ToNot(HaveOccurred())
		})

		It("creates a sha1 digest", func() {
			err := json.Unmarshal([]byte(`"sha1:07e1306432667f916639d47481edc4f2ca456454"`), &digest)
			Expect(err).ToNot(HaveOccurred())
			Expect(digest.Algorithm().Name()).To(Equal(DigestAlgorithmSHA1.Name()))
			Expect(digest.String()).To(Equal("07e1306432667f916639d47481edc4f2ca456454"))
		})

		It("creates a sha256 digest", func() {
			err := json.Unmarshal([]byte(`"sha256:b1e66f505465c28d705cf587b041a6506cfe749f7aa4159d8a3f45cc53f1fb23"`), &digest)
			Expect(err).ToNot(HaveOccurred())
			Expect(digest.Algorithm().Name()).To(Equal(DigestAlgorithmSHA256.Name()))
			Expect(digest.String()).To(Equal("sha256:b1e66f505465c28d705cf587b041a6506cfe749f7aa4159d8a3f45cc53f1fb23"))
		})

		It("creates a sha512 digest", func() {
			err := json.Unmarshal([]byte(`"sha512:6f06a0c6c3827d827145b077cd8c8b7a15c75eb2bed809569296e6502ef0872c8e7ef91307a6994fcd2be235d3c41e09bfe1b6023df45697d88111df4349d64a"`), &digest)
			Expect(err).ToNot(HaveOccurred())
			Expect(digest.Algorithm().Name()).To(Equal(DigestAlgorithmSHA512.Name()))
			Expect(digest.String()).To(Equal("sha512:6f06a0c6c3827d827145b077cd8c8b7a15c75eb2bed809569296e6502ef0872c8e7ef91307a6994fcd2be235d3c41e09bfe1b6023df45697d88111df4349d64a"))
		})

		It("creates a sha1 digest when algo is not specified", func() {
			err := json.Unmarshal([]byte(`"07e1306432667f916639d47481edc4f2ca456454"`), &digest)
			Expect(err).ToNot(HaveOccurred())
			Expect(digest.Algorithm().Name()).To(Equal(DigestAlgorithmSHA1.Name()))
			Expect(digest.String()).To(Equal("07e1306432667f916639d47481edc4f2ca456454"))
		})

		It("retains unknown algos", func() {
			err := json.Unmarshal([]byte(`"unknown1:val1;unknown2:val2"`), &digest)
			Expect(err).ToNot(HaveOccurred())
			Expect(digest.Algorithm().Name()).To(Equal("unknown1"))
			Expect(digest.String()).To(Equal("unknown1:val1"))
		})

		It("does not error when the json contains a valid digest and an unknown digest", func() {
			err := json.Unmarshal([]byte(`"unknown1:val1;sha256:1bf4b70c96b9d4e8f473ac6b7e6b5b965ab3497287a86eb2ed1b263287c78038"`), &digest)
			Expect(err).ToNot(HaveOccurred())

			Expect(digest.Algorithm().Name()).To(Equal(DigestAlgorithmSHA256.Name()))
			Expect(digest.String()).To(Equal("sha256:1bf4b70c96b9d4e8f473ac6b7e6b5b965ab3497287a86eb2ed1b263287c78038"))
			Expect(digest.Verify(strings.NewReader("content to be verified"))).ToNot(HaveOccurred())
		})

		It("returns an error if the JSON does not contain any digests", func() {
			err := json.Unmarshal([]byte(`""`), &digest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No recognizable digest algorithm found"))
		})

		It("returns an error if the JSON contains only semicolon", func() {
			err := json.Unmarshal([]byte(`";"`), &digest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No recognizable digest algorithm found"))
		})
	})
})
