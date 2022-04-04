package cmd_test

import (
	"crypto/tls"

	"github.com/onsi/gomega/ghttp"
)

func BuildSSLServer() (*ghttp.Server, string) {
	server := ghttp.NewUnstartedServer()

	server.HTTPTestServer.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server.HTTPTestServer.StartTLS()

	return server, validCACert
}
