package registry

import (
	"net"
	"net/http"
	"net/url"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Server interface {
	Start() error
	Stop() error
}

type server struct {
	url      *url.URL
	listener net.Listener
	logger   boshlog.Logger
	logTag   string
}

func NewServer(endpoint string, logger boshlog.Logger) Server {
	url, _ := url.Parse(endpoint)
	return &server{
		url:    url,
		logger: logger,
		logTag: "registryServer",
	}
}

func (s *server) Start() error {
	s.logger.Debug(s.logTag, "Starting registry server")

	var err error
	s.listener, err = net.Listen("tcp", s.url.Host)
	if err != nil {
		return bosherr.WrapError(err, "Starting registry listener")
	}

	httpServer := http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	username := s.url.User.Username()
	password, _ := s.url.User.Password()
	registry := NewRegistry()
	instanceHandler := NewInstanceHandler(username, password, registry, s.logger)
	mux.HandleFunc("/instances/", instanceHandler.HandleFunc)

	return httpServer.Serve(s.listener)
}

func (s *server) Stop() error {
	s.logger.Debug(s.logTag, "Stopping registry server")
	err := s.listener.Close()
	if err != nil {
		return bosherr.WrapError(err, "Stopping registry server")
	}

	return nil
}
