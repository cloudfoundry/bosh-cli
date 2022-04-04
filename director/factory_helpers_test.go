package director_test

import (
	"crypto/tls"

	. "github.com/cloudfoundry/bosh-cli/director"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
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
