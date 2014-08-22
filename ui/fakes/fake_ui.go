package fakes

type FakeUI struct {
	Said   []string
	Errors []string
}

func (ui *FakeUI) Sayln(message string) {
	ui.Said = append(ui.Said, message)
}

func (ui *FakeUI) Error(message string) {
	ui.Errors = append(ui.Errors, message)
}
