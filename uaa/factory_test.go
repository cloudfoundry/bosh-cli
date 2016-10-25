package uaa_test

import (
	"crypto/tls"
	"net/http"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/uaa"
)

var _ = Describe("Factory", func() {
	Describe("New", func() {
		It("returns error if config is invalid", func() {
			_, err := NewFactory(boshlog.NewLogger(boshlog.LevelNone)).New(Config{})
			Expect(err).To(HaveOccurred())
		})

		It("UAA returns error if TLS cannot be verified", func() {
			server := ghttp.NewTLSServer()
			defer server.Close()

			config, err := NewConfigFromURL(server.URL())
			Expect(err).ToNot(HaveOccurred())

			config.Client = "client"
			config.ClientSecret = "client-secret"

			logger := boshlog.NewLogger(boshlog.LevelNone)

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

			logger := boshlog.NewLogger(boshlog.LevelNone)

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
		})
	})
})
