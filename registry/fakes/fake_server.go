package fakes

type FakeServer struct {
	StartErr error
	StopErr  error

	ReceivedActions []string
}

func NewFakeServer() *FakeServer {
	return &FakeServer{
		ReceivedActions: []string{},
	}
}

func (s *FakeServer) Start() error {
	s.ReceivedActions = append(s.ReceivedActions, "Start")
	return s.StartErr
}

func (s *FakeServer) Stop() error {
	s.ReceivedActions = append(s.ReceivedActions, "Stop")
	return s.StopErr
}
