package sshtunnel

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Options struct {
	Host       string
	Port       int
	User       string
	PrivateKey string
	Password   string

	LocalForwardPort  int
	RemoteForwardPort int
}

func (o Options) IsEmpty() bool {
	return o == Options{}
}

type Factory interface {
	NewSSHTunnel(Options) SSHTunnel
}

type factory struct {
	logger boshlog.Logger
}

func NewFactory(logger boshlog.Logger) Factory {
	return &factory{
		logger: logger,
	}
}

func (s *factory) NewSSHTunnel(options Options) SSHTunnel {
	return &sshTunnel{
		startDialMaxTries: 300,
		startDialDelay:    500 * time.Millisecond,
		options:           options,
		logger:            s.logger,
		logTag:            "sshTunnel",
	}
}
