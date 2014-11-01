package registry

import (
	"fmt"
	"net"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Server interface {
	Start(string, string, string, int, chan error) error
	Stop() error
}

type server struct {
	listener net.Listener
	logger   boshlog.Logger
	logTag   string
}

func NewServer(logger boshlog.Logger) Server {
	return &server{
		logger: logger,
		logTag: "registryServer",
	}
}

func (s *server) Start(username string, password string, host string, port int, readyErrCh chan error) error {
	s.logger.Debug(s.logTag, "Starting registry server at %s:%d", host, port)
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		readyErrCh <- bosherr.WrapError(err, "Starting registry listener")
		return nil
	}

	readyErrCh <- nil

	httpServer := http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

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
