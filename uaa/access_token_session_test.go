package uaa_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-cli/uaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/cmd/config/configfakes"

	fakeuaa "github.com/cloudfoundry/bosh-cli/uaa/uaafakes"
)

var _ = Describe("AccessTokenSession", func() {
	var (
		uaa       *fakeuaa.FakeUAA
		initToken *fakeuaa.FakeRefreshableAccessToken
		config    *configfakes.FakeConfig
		sess      *AccessTokenSession
	)

	BeforeEach(func() {
		uaa = &fakeuaa.FakeUAA{}
	})

	Describe("TokenFunc", func() {
		BeforeEach(func() {
			initToken = &fakeuaa.FakeRefreshableAccessToken{}
			config = &configfakes.FakeConfig{}
			sess = NewAccessTokenSession(uaa, initToken, config, "url")
			initToken.IsValidReturns(false)
		})

		Context("when initial token is invalid", func() {
			It("returns an auth header with a new token and updates config", func() {
				token := &fakeuaa.FakeRefreshableAccessToken{
					TypeStub:         func() string { return "type1" },
					ValueStub:        func() string { return "value1" },
					RefreshValueStub: func() string { return "refresh-value1" },
				}
				uaa.RefreshTokenGrantReturns(token, nil)

				header, err := sess.TokenFunc(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(header).To(Equal("type1 value1"))
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(1))
				env, updatedToken := config.UpdateConfigWithTokenArgsForCall(0)
				Expect(env).To(Equal("url"))
				Expect(updatedToken).To(Equal(token))
			})

			It("returns an error if obtaining token fails", func() {
				token := &fakeuaa.FakeRefreshableAccessToken{}
				uaa.RefreshTokenGrantReturns(token, errors.New("fake-err"))

				_, err := sess.TokenFunc(false)
				Expect(err).To(HaveOccurred())
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(0))
			})
		})

		Context("when retrying is set", func() {
			It("returns an auth header with a new token", func() {
				token := &fakeuaa.FakeRefreshableAccessToken{
					TypeStub:         func() string { return "type1" },
					ValueStub:        func() string { return "value1" },
					RefreshValueStub: func() string { return "refresh-value1" },
				}
				uaa.RefreshTokenGrantReturns(token, nil)

				header, err := sess.TokenFunc(true)
				Expect(err).ToNot(HaveOccurred())
				Expect(header).To(Equal("type1 value1"))
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(1))
				env, updatedToken := config.UpdateConfigWithTokenArgsForCall(0)
				Expect(env).To(Equal("url"))
				Expect(updatedToken).To(Equal(token))
			})

			It("returns an error if obtaining token fails", func() {
				token := &fakeuaa.FakeRefreshableAccessToken{}
				uaa.RefreshTokenGrantReturns(token, errors.New("fake-err"))

				_, err := sess.TokenFunc(true)
				Expect(err).To(HaveOccurred())
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(0))
			})
		})

		Context("when saving the config fails", func() {
			It("returns an error", func() {
				config.UpdateConfigWithTokenReturns(errors.New("fake-err"))

				_, err := sess.TokenFunc(true)
				Expect(err).To(MatchError("fake-err"))
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(1))
			})
		})
		Context("when retrying is not set", func() {
			It("returns an auth header with a new token", func() {
				token := &fakeuaa.FakeAccessToken{
					TypeStub:  func() string { return "type1" },
					ValueStub: func() string { return "value1" },
				}
				uaa.RefreshTokenGrantReturns(token, nil)

				header, err := sess.TokenFunc(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(header).To(Equal("type1 value1"))
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(1))
				env, updatedToken := config.UpdateConfigWithTokenArgsForCall(0)
				Expect(env).To(Equal("url"))
				Expect(updatedToken).To(Equal(token))
			})

			It("returns an error if obtaining first token fails", func() {
				token := &fakeuaa.FakeAccessToken{}
				uaa.RefreshTokenGrantReturns(token, errors.New("fake-err"))

				_, err := sess.TokenFunc(false)
				Expect(err).To(HaveOccurred())
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(0))
			})
		})

		Context("when not refreshable", func() {
			It("returns an error", func() {
				token := &fakeuaa.FakeAccessToken{}

				uaa.RefreshTokenGrantReturns(token, errors.New("fake-err"))
				sess = NewAccessTokenSession(uaa, token, config, "url")

				_, err := sess.TokenFunc(true)
				Expect(err).To(MatchError("not a refresh token"))
				Expect(config.UpdateConfigWithTokenCallCount()).To(Equal(0))
			})
		})
	})
})
