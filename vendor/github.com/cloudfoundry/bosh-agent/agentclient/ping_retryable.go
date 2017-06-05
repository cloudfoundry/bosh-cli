package agentclient

import (
	"crypto/x509"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
)

type pingRetryable struct {
	agentClient AgentClient
}

func NewPingRetryable(agentClient AgentClient) boshretry.Retryable {
	return &pingRetryable{
		agentClient: agentClient,
	}
}

func (r *pingRetryable) Attempt() (bool, error) {
	_, err := r.agentClient.Ping()

	if err != nil {
		for {
			if innerErr, ok := err.(bosherr.ComplexError); ok {
				err = innerErr.Cause
			} else {
				break
			}
		}
		if _, ok := err.(x509.CertificateInvalidError); ok {
			return false, err
		}
	}

	return true, err
}
