package uaa_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/uaa"
)

var _ = Describe("AccessTokenImpl", func() {
	var token AccessToken

	BeforeEach(func() {
		token = NewAccessToken("token-type", "token-value")
	})

	Describe("IsValid", func() {
		It("is invalid if it has an empty type", func() {
			token = NewAccessToken("", "token-value")
			Expect(token.IsValid()).To(BeFalse())
		})

		It("is invalid if it has an empty value", func() {
			token = NewAccessToken("token-type", "")
			Expect(token.IsValid()).To(BeFalse())
		})

		It("is valid if it has both type and value", func() {
			token = NewAccessToken("token-type", "token-value")
			Expect(token.IsValid()).To(BeTrue())
		})
	})

	Describe("Type", func() {
		It("returns", func() {
			Expect(token.Type()).To(Equal("token-type"))
		})
	})

	Describe("Value", func() {
		It("returns", func() {
			Expect(token.Value()).To(Equal("token-value"))
		})
	})
})

var _ = Describe("RefreshableAccessTokenImpl", func() {
	var token RefreshableAccessToken

	BeforeEach(func() {
		token = NewRefreshableAccessToken("token-type", "token-value", "refresh-value")
	})

	Describe("Type", func() {
		It("returns", func() {
			Expect(token.Type()).To(Equal("token-type"))
		})
	})

	Describe("Value", func() {
		It("returns", func() {
			Expect(token.Value()).To(Equal("token-value"))
		})
	})

	Describe("RefreshValue", func() {
		It("returns", func() {
			Expect(token.RefreshValue()).To(Equal("refresh-value"))
		})
	})

	Describe("IsValid", func() {
		It("is invalid if it has an empty type", func() {
			token = NewRefreshableAccessToken("", "token-value", "refresh-value")
			Expect(token.IsValid()).To(BeFalse())
		})

		It("is invalid if it has an empty value", func() {
			token = NewRefreshableAccessToken("token-type", "", "refresh-value")
			Expect(token.IsValid()).To(BeFalse())
		})

		It("is valid if it has both type and value", func() {
			token = NewRefreshableAccessToken("token-type", "token-value", "refresh-value")
			Expect(token.IsValid()).To(BeTrue())
		})
	})

	It("panics if refresh value is empty", func() {
		Expect(func() { NewRefreshableAccessToken("access-token-type", "access-token", "") }).To(Panic())
	})
})

var _ = Describe("NewTokenInfoFromValue", func() {
	It("returns parsed token", func() {
		info, err := NewTokenInfoFromValue("seg.eyJ1c2VyX25hbWUiOiJhZG1pbiIsInNjb3BlIjpbIm9wZW5pZCIsImJvc2guYWRtaW4iXSwiZXhwIjoxMjN9.seg")
		Expect(err).ToNot(HaveOccurred())
		Expect(info).To(Equal(TokenInfo{
			Username:  "admin",
			Scopes:    []string{"openid", "bosh.admin"},
			ExpiredAt: 123,
		}))
	})

	It("returns an error if token doesnt have 3 segments", func() {
		_, err := NewTokenInfoFromValue("seg")
		Expect(err).To(Equal(errors.New("Expected token value to have 3 segments")))

		_, err = NewTokenInfoFromValue("seg.seg")
		Expect(err).To(Equal(errors.New("Expected token value to have 3 segments")))

		_, err = NewTokenInfoFromValue("seg.seg.seg.seg")
		Expect(err).To(Equal(errors.New("Expected token value to have 3 segments")))
	})

	It("returns an error if token's 2nd segment cannot be decoded", func() {
		_, err := NewTokenInfoFromValue("seg.eyJrZXkiOiJ2YWx1ZXoifQ==.seg")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Decoding token info"))
	})

	It("returns an error if token's 2nd segment cannot unmarshaled", func() {
		_, err := NewTokenInfoFromValue("seg.a2V5.seg")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unmarshaling token info"))
	})
})
