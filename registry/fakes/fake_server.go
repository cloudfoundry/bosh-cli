package fakes

type FakeServer struct {
	StartInput StartInput
	readyChErr error
	startErr   error

	stopErr error

	ReceivedActions []string
}

type StartInput struct {
	Username string
	Password string
	Host     string
	Port     int
}

func NewFakeServer() *FakeServer {
	return &FakeServer{
		ReceivedActions: []string{},
		StartInput:      StartInput{},
	}
}

func (s *FakeServer) Start(username string, password string, host string, port int, readyErrCh chan error) error {
	s.StartInput = StartInput{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	}
	s.ReceivedActions = append(s.ReceivedActions, "Start")

	readyErrCh <- s.readyChErr
	return s.startErr
}

func (s *FakeServer) Stop() error {
	s.ReceivedActions = append(s.ReceivedActions, "Stop")
	return s.stopErr
}

func (s *FakeServer) SetStartBehavior(readyChErr, startErr error) {
	s.readyChErr = readyChErr
	s.startErr = startErr
}
