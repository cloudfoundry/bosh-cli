package ui

import (
	"fmt"

	"github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type indentingUI struct {
	parent UI
}

func NewIndentingUI(parent UI) UI {
	return &indentingUI{parent: parent}
}

func (ui *indentingUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.parent.ErrorLinef("  %s", fmt.Sprintf(pattern, args...))
}

func (ui *indentingUI) PrintLinef(pattern string, args ...interface{}) {
	ui.parent.PrintLinef("  %s", fmt.Sprintf(pattern, args...))
}

func (ui *indentingUI) BeginLinef(pattern string, args ...interface{}) {
	ui.parent.BeginLinef("  %s", fmt.Sprintf(pattern, args...))
}

func (ui *indentingUI) EndLinef(pattern string, args ...interface{}) {
	ui.parent.EndLinef(pattern, args...)
}

func (ui *indentingUI) PrintBlock(block []byte) {
	ui.parent.PrintBlock(block)
}

func (ui *indentingUI) PrintErrorBlock(block string) {
	ui.parent.PrintErrorBlock(block)
}

func (ui *indentingUI) PrintTable(table table.Table) {
	ui.parent.PrintTable(table)
}

func (ui *indentingUI) PrintTableFiltered(table table.Table, filterHeader []table.Header) {
	ui.parent.PrintTableFiltered(table, filterHeader)
}

func (ui *indentingUI) AskForText(label string) (string, error) {
	return ui.parent.AskForText(label)
}

func (ui *indentingUI) AskForTextWithDefaultValue(label, defaultValue string) (string, error) {
	return ui.parent.AskForTextWithDefaultValue(label, defaultValue)
}

func (ui *indentingUI) AskForChoice(label string, options []string) (int, error) {
	return ui.parent.AskForChoice(label, options)
}

func (ui *indentingUI) AskForPassword(label string) (string, error) {
	return ui.parent.AskForPassword(label)
}

func (ui *indentingUI) AskForConfirmation() error {
	return ui.parent.AskForConfirmation()
}

func (ui *indentingUI) AskForConfirmationWithLabel(label string) error {
	return ui.parent.AskForConfirmationWithLabel(label)
}

func (ui *indentingUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *indentingUI) Flush() {
	ui.parent.Flush()
}
