package ui

import (
	"sync"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type lockingUI struct {
	parent     UI
	printMutex *sync.Mutex
}

func NewLockingUI(parent UI) UI {
	return &lockingUI{parent: parent, printMutex: &sync.Mutex{}}
}

func (ui *lockingUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.printMutex.Lock()
	defer ui.printMutex.Unlock()
	ui.parent.ErrorLinef(pattern, args...)
}

func (ui *lockingUI) PrintLinef(pattern string, args ...interface{}) {
	ui.printMutex.Lock()
	defer ui.printMutex.Unlock()
	ui.parent.PrintLinef(pattern, args...)
}

func (ui *lockingUI) BeginLinef(pattern string, args ...interface{}) {
	ui.printMutex.Lock()
	defer ui.printMutex.Unlock()
	ui.parent.BeginLinef(pattern, args...)
}

func (ui *lockingUI) EndLinef(pattern string, args ...interface{}) {
	ui.printMutex.Lock()
	defer ui.printMutex.Unlock()
	ui.parent.EndLinef(pattern, args...)
}

func (ui *lockingUI) PrintBlock(block []byte) {
	ui.printMutex.Lock()
	defer ui.printMutex.Unlock()
	ui.parent.PrintBlock(block)
}

func (ui *lockingUI) PrintErrorBlock(block string) {
	ui.printMutex.Lock()
	defer ui.printMutex.Unlock()
	ui.parent.PrintErrorBlock(block)
}

func (ui *lockingUI) PrintTable(table Table) {
	ui.printMutex.Lock()
	defer ui.printMutex.Unlock()
	ui.parent.PrintTable(table)
}

func (ui *lockingUI) AskForText(label string) (string, error) {
	return ui.parent.AskForText(label)
}

func (ui *lockingUI) AskForChoice(label string, options []string) (int, error) {
	return ui.parent.AskForChoice(label, options)
}

func (ui *lockingUI) AskForPassword(label string) (string, error) {
	return ui.parent.AskForPassword(label)
}

func (ui *lockingUI) AskForConfirmation() error {
	return ui.parent.AskForConfirmation()
}

func (ui *lockingUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *lockingUI) Flush() {
	ui.parent.Flush()
}
