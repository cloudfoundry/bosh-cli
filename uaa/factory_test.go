package uaa_test

import (
	"crypto/tls"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/uaa"
	"github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
)

var _ = Describe("Factory", func() {
	Describe("New", func() {
		var logger *loggerfakes.FakeLogger

		BeforeEach(func() {
			logger = &loggerfakes.FakeLogger{}
		})

		It("returns error if config is invalid", func() {
			_, err := NewFactory(logger).New(Config{})
			Expect(err).To(HaveOccurred())
		})

		It("UAA returns error if TLS cannot be verified", func() {
			server := ghttp.NewTLSServer()
			defer server.Close()

			config, err := NewConfigFromURL(server.URL())
			Expect(err).ToNot(HaveOccurred())

			config.Client = "client"
			config.ClientSecret = "client-secret"

			uaa, err := NewFactory(logger).New(config)
			Expect(err).ToNot(HaveOccurred())

			_, err = uaa.ClientCredentialsGrant()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("x509: certificate signed by unknown authority"))
		})

		It("UAA succeeds making a request with client creds if TLS can be verified", func() {
			server := ghttp.NewUnstartedServer()

			cert, err := tls.X509KeyPair(validCert, validKey)
			Expect(err).ToNot(HaveOccurred())

			server.HTTPTestServer.TLS = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			server.HTTPTestServer.StartTLS()

			config, err := NewConfigFromURL(server.URL())
			Expect(err).ToNot(HaveOccurred())

			config.Client = "client"
			config.ClientSecret = "client-secret"
			config.CACert = validCACert

			uaa, err := NewFactory(logger).New(config)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token", "grant_type=client_credentials"),
					ghttp.VerifyBasicAuth("client", "client-secret"),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			_, err = uaa.ClientCredentialsGrant()
			Expect(err).ToNot(HaveOccurred())
			_, _, args := logger.DebugArgsForCall(1)
			Expect(args[0]).To(ContainSubstring("/token?grant_type=<redacted>"))
		})

		Context("when the server url has a context path", func() {
			It("properly follows that path", func() {
				server := ghttp.NewUnstartedServer()

				cert, err := tls.X509KeyPair(validCert, validKey)
				Expect(err).ToNot(HaveOccurred())

				server.HTTPTestServer.TLS = &tls.Config{
					Certificates: []tls.Certificate{cert},
				}

				server.HTTPTestServer.StartTLS()

				config, err := NewConfigFromURL(fmt.Sprintf("%s/test_path", server.URL()))
				Expect(err).ToNot(HaveOccurred())

				config.Client = "client"
				config.ClientSecret = "client-secret"
				config.CACert = validCACert

				uaa, err := NewFactory(logger).New(config)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/test_path/oauth/token", "grant_type=client_credentials"),
						ghttp.VerifyBasicAuth("client", "client-secret"),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				_, err = uaa.ClientCredentialsGrant()
				Expect(err).ToNot(HaveOccurred())

			})
		})
	})
})
