package fakes

type FakeServer struct {
	StartInput StartInput
	StartErr   error

	StopErr error

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

func (s *FakeServer) Start(username string, password string, host string, port int, readyCh chan struct{}) error {
	s.StartInput = StartInput{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
	}
	s.ReceivedActions = append(s.ReceivedActions, "Start")

	readyCh <- struct{}{}
	return s.StartErr
}

func (s *FakeServer) Stop() error {
	s.ReceivedActions = append(s.ReceivedActions, "Stop")
	return s.StopErr
}
