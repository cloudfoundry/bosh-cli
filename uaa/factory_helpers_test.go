package uaa_test

import (
	"crypto/tls"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/uaa"
)

func BuildServer() (UAA, *ghttp.Server) {
	server := ghttp.NewUnstartedServer()

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

	return uaa, server
}
