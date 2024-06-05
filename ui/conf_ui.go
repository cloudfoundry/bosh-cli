package ui

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type ConfUI struct {
	parent      UI
	isTTY       bool
	logger      boshlog.Logger
	showColumns []Header
}

func NewConfUI(logger boshlog.Logger) *ConfUI {
	var ui UI

	writerUI := NewConsoleUI(logger)
	ui = NewPaddingUI(writerUI)

	return &ConfUI{
		parent: ui,
		isTTY:  writerUI.IsTTY(),
		logger: logger,
	}
}

func NewWrappingConfUI(parent UI, logger boshlog.Logger) *ConfUI {
	return &ConfUI{
		parent: parent,
		isTTY:  true,
		logger: logger,
	}
}

func (ui *ConfUI) EnableTTY(force bool) {
	if !ui.isTTY && !force {
		ui.parent = NewNonTTYUI(ui.parent)
	}
}

func (ui *ConfUI) EnableColor() {
	ui.parent = NewColorUI(ui.parent)
}

func (ui *ConfUI) EnableJSON() {
	ui.parent = NewJSONUI(ui.parent, ui.logger)
}

func (ui *ConfUI) ShowColumns(columns []Header) {
	ui.showColumns = columns
}

func (ui *ConfUI) EnableNonInteractive() {
	ui.parent = NewNonInteractiveUI(ui.parent)
}

func (ui *ConfUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.parent.ErrorLinef(pattern, args...)
}

func (ui *ConfUI) PrintLinef(pattern string, args ...interface{}) {
	ui.parent.PrintLinef(pattern, args...)
}

func (ui *ConfUI) BeginLinef(pattern string, args ...interface{}) {
	ui.parent.BeginLinef(pattern, args...)
}

func (ui *ConfUI) EndLinef(pattern string, args ...interface{}) {
	ui.parent.EndLinef(pattern, args...)
}

func (ui *ConfUI) PrintBlock(block []byte) {
	ui.parent.PrintBlock(block)
}

func (ui *ConfUI) PrintErrorBlock(block string) {
	ui.parent.PrintErrorBlock(block)
}

func (ui *ConfUI) PrintTable(table Table) {
	if len(ui.showColumns) > 0 {
		err := table.SetColumnVisibility(ui.showColumns)
		if err != nil {
			ui.parent.PrintErrorBlock(err.Error())
		}
	}

	ui.parent.PrintTable(table)
}

func (ui *ConfUI) PrintTableFiltered(table Table, filterHeader []Header) {
	if len(ui.showColumns) > 0 {
		err := table.SetColumnVisibilityFiltered(ui.showColumns, filterHeader)
		if err != nil {
			ui.parent.PrintErrorBlock(err.Error())
		}
	}

	ui.parent.PrintTableFiltered(table, filterHeader)
}

func (ui *ConfUI) AskForText(label string) (string, error) {
	return ui.parent.AskForText(label)
}

func (ui *ConfUI) AskForTextWithDefaultValue(label, defaultValue string) (string, error) {
	return ui.parent.AskForTextWithDefaultValue(label, defaultValue)
}

func (ui *ConfUI) AskForChoice(label string, options []string) (int, error) {
	return ui.parent.AskForChoice(label, options)
}

func (ui *ConfUI) AskForPassword(label string) (string, error) {
	return ui.parent.AskForPassword(label)
}

func (ui *ConfUI) AskForConfirmation() error {
	return ui.parent.AskForConfirmation()
}

func (ui *ConfUI) AskForConfirmationWithLabel(label string) error {
	return ui.parent.AskForConfirmationWithLabel(label)
}

func (ui *ConfUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *ConfUI) Flush() {
	ui.parent.Flush()
}
