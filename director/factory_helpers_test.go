package director_test

import (
	"crypto/tls"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/v7/director"
)

func BuildServer() (Director, *ghttp.Server) {
	server := ghttp.NewUnstartedServer()

	server.HTTPTestServer.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server.HTTPTestServer.StartTLS()

	factoryConfig, err := NewConfigFromURL(server.URL())
	Expect(err).ToNot(HaveOccurred())

	factoryConfig.Client = "username"
	factoryConfig.ClientSecret = "password"
	factoryConfig.CACert = validCACert

	logger := boshlog.NewLogger(boshlog.LevelNone)
	taskReporter := NewNoopTaskReporter()
	fileReporter := NewNoopFileReporter()

	director, err := NewFactory(logger).New(factoryConfig, taskReporter, fileReporter)
	Expect(err).ToNot(HaveOccurred())

	return director, server
}
