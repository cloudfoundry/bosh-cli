// Package grpcrunner simplifies starting GRPC servers with ifrit.
package grpcrunner

import (
	"net"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/tedsuo/ifrit"
	"google.golang.org/grpc"
)

type grpcServer struct {
	logger lager.Logger

	listenAddr   string
	registerFunc func(*grpc.Server)
	options      []grpc.ServerOption
}

// New creates a new runner for a GRPC server that can be started with ifrit.
// The registerFunc can be used to register your GPRC services with the running
// server. The options are passed through directly to the server.
func New(
	logger lager.Logger,
	listenAddr string,
	registerFunc func(*grpc.Server),
	options ...grpc.ServerOption,
) ifrit.Runner {
	return &grpcServer{
		logger:       logger,
		listenAddr:   listenAddr,
		registerFunc: registerFunc,
		options:      options,
	}
}

func (s *grpcServer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := s.logger.Session("grpc-server", lager.Data{
		"addr": s.listenAddr,
	})

	lis, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(s.options...)
	s.registerFunc(grpcServer)

	errCh := make(chan error, 1)
	go func() {
		errCh <- grpcServer.Serve(lis)
	}()

	close(ready)

	logger.Info("started")

	select {
	case err = <-errCh:
		return err
	case <-signals:
		logger.Info("signalled")

		grpcServer.GracefulStop()
	}

	logger.Info("exited")

	return nil
}
