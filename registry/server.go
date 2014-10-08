package registry

import (
	"fmt"
	"net"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Server interface {
	Start() error
	Stop() error
}

type server struct {
	username string
	password string
	host     string
	port     int
	listener net.Listener
	logger   boshlog.Logger
	logTag   string
}

func NewServer(username string, password string, host string, port int, logger boshlog.Logger) Server {
	return &server{
		username: username,
		password: password,
		host:     host,
		port:     port,
		logger:   logger,
		logTag:   "registryServer",
	}
}

func (s *server) Start() error {
	s.logger.Debug(s.logTag, "Starting registry server")

	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return bosherr.WrapError(err, "Starting registry listener")
	}

	httpServer := http.Server{}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	registry := NewRegistry()
	instanceHandler := NewInstanceHandler(s.username, s.password, registry, s.logger)
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
