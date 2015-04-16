package sshtunnel

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
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
	timeService := boshtime.NewConcreteService()
	return &sshTunnel{
		connectionRefusedTimeout: 5 * time.Minute,
		authFailureTimeout:       2 * time.Minute,
		startDialDelay:           500 * time.Millisecond,
		timeService:              timeService,
		options:                  options,
		logger:                   s.logger,
		logTag:                   "sshTunnel",
	}
}
