package fakes

type FakeTunnel struct {
	startOutput *startOutput
	Started     bool
}

type startOutput struct {
	ReadyChOutput struct{}
	ErrChOutput   error
}

func NewFakeTunnel() *FakeTunnel {
	return &FakeTunnel{}
}

func (s *FakeTunnel) Start(readyCh chan<- struct{}, errCh chan<- error) {
	s.Started = true

	if s.startOutput != nil {
		readyCh <- s.startOutput.ReadyChOutput
		errCh <- s.startOutput.ErrChOutput
	}
}

func (s *FakeTunnel) Stop() error {
	return nil
}

func (s *FakeTunnel) SetStartBehavior(readyChOutput struct{}, errChOutput error) {
	s.startOutput = &startOutput{
		ReadyChOutput: readyChOutput,
		ErrChOutput:   errChOutput,
	}
}
