package http

//go:generate mockgen -source=agent_client_factory.go -package=mocks -destination=mocks/mocks.go

import (
	"net/http"
	"time"

	"github.com/cloudfoundry/bosh-agent/agentclient"
	"github.com/cloudfoundry/bosh-utils/crypto"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type AgentClientFactory interface {
	NewAgentClient(directorID, mbusURL, caCert string) (agentclient.AgentClient, error)
}

type agentClientFactory struct {
	getTaskDelay  time.Duration
	clientFactory httpclient.ClientFactory
	logger        boshlog.Logger
}

func NewAgentClientFactory(
	getTaskDelay time.Duration,
	clientFactory httpclient.ClientFactory,
	logger boshlog.Logger,
) AgentClientFactory {
	return &agentClientFactory{
		getTaskDelay:  getTaskDelay,
		logger:        logger,
		clientFactory: clientFactory,
	}
}

func (f *agentClientFactory) NewAgentClient(directorID, mbusURL, caCert string) (agentclient.AgentClient, error) {
	var client *http.Client
	if caCert != "" {
		caCertPool, err := crypto.CertPoolFromPEM([]byte(caCert))
		if err != nil {
			return nil, err
		}
		client = f.clientFactory.CreateDefaultClient(caCertPool)
	} else {
		client = f.clientFactory.CreateDefaultClientInsecureSkipVerify()
	}

	httpClient := httpclient.NewHTTPClient(client, f.logger)
	return NewAgentClient(mbusURL, directorID, f.getTaskDelay, 10, httpClient, f.logger), nil
}
