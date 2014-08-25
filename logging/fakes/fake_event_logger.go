package fakes

type FakeEventLogger struct {
	LoggedEvents      []string
	StartedGroup      string
	FinishGroupCalled bool
}

func NewFakeEventLogger() *FakeEventLogger {
	return &FakeEventLogger{}
}

func (fl *FakeEventLogger) TrackAndLog(event string, f func() error) error {
	err := f()
	fl.LoggedEvents = append(fl.LoggedEvents, event)
	return err
}

func (fl *FakeEventLogger) StartGroup(group string) error {
	fl.StartedGroup = group
	return nil
}

func (fl *FakeEventLogger) FinishGroup() error {
	fl.FinishGroupCalled = true
	return nil
}
