package eventlogger

type stage struct {
	name        string
	eventLogger EventLogger
}

type Stage interface {
	NewStep(string) Step
	PerformStep(string, func() error) error
	Name() string
	Start()
	Finish()
}

func NewStage(name string, eventLogger EventLogger) Stage {
	return &stage{
		name:        name,
		eventLogger: eventLogger,
	}
}

func (s *stage) NewStep(stepName string) Step {
	step := &step{
		name:        stepName,
		stage:       s,
		eventLogger: s.eventLogger,
	}

	return step
}

func (s *stage) PerformStep(stepName string, stepFunc func() error) error {
	step := s.NewStep(stepName)
	step.Start()
	err := stepFunc()
	if err != nil {
		step.Fail(err.Error())
		return err
	}
	step.Finish()
	return nil
}

func (s *stage) Name() string {
	return s.name
}

func (s *stage) Start() {
	s.eventLogger.StartStage(s.name)
}

func (s *stage) Finish() {
	s.eventLogger.FinishStage(s.name)
}
