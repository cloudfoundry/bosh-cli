package integration_test

import (
	"crypto/tls"

	"github.com/onsi/gomega/ghttp"
)

func BuildHTTPSServer() (string, *ghttp.Server) {
	server := ghttp.NewUnstartedServer()

	server.HTTPTestServer.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server.HTTPTestServer.StartTLS()

	return validCACert, server
}
