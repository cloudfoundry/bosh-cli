package eventlogger

type step struct {
	name        string
	stage       Stage
	index       int
	eventLogger EventLogger
}

type Step interface {
	Start()
	Finish()
	Skip(string)
	Fail(string)
}

func (s *step) Start() {
	event := Event{
		Stage: s.stage.Name(),
		Task:  s.name,
		State: Started,
	}
	s.eventLogger.AddEvent(event)
}

func (s *step) Finish() {
	event := Event{
		Stage: s.stage.Name(),
		Task:  s.name,
		State: Finished,
	}
	s.eventLogger.AddEvent(event)
}

func (s *step) Skip(message string) {
	event := Event{
		Stage:   s.stage.Name(),
		Task:    s.name,
		State:   Skipped,
		Message: message,
	}
	s.eventLogger.AddEvent(event)
}

func (s *step) Fail(message string) {
	event := Event{
		Stage:   s.stage.Name(),
		Task:    s.name,
		State:   Failed,
		Message: message,
	}
	s.eventLogger.AddEvent(event)
}
