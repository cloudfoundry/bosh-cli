package eventlogger

type stage struct {
	name         string
	totalSteps   int
	eventLogger  EventLogger
	currentIndex int
}

type Stage interface {
	NewStep(string) Step
	Name() string
	TotalSteps() int
}

func NewStage(name string, totalSteps int, eventLogger EventLogger) Stage {
	return &stage{
		name:        name,
		totalSteps:  totalSteps,
		eventLogger: eventLogger,
	}
}

func (s *stage) NewStep(stepName string) Step {
	s.currentIndex++
	step := &step{
		name:        stepName,
		stage:       s,
		index:       s.currentIndex,
		eventLogger: s.eventLogger,
	}

	return step
}

func (s *stage) Name() string {
	return s.name
}

func (s *stage) TotalSteps() int {
	return s.totalSteps
}
