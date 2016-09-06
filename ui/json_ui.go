package ui

import (
	"encoding/json"
	"fmt"
	"reflect"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

type jsonUI struct {
	parent UI
	uiResp uiResp

	logTag string
	logger boshlog.Logger
}

type uiResp struct {
	Tables []tableResp
	Blocks []string
	Lines  []string
}

type tableResp struct {
	Content string
	Header  []string
	Rows    [][]string
	Notes   []string
}

func NewJSONUI(parent UI, logger boshlog.Logger) UI {
	return &jsonUI{parent: parent, logTag: "JSONUI", logger: logger}
}

func (ui *jsonUI) ErrorLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *jsonUI) PrintLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *jsonUI) BeginLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *jsonUI) EndLinef(pattern string, args ...interface{}) {
	ui.addLine(pattern, args)
}

func (ui *jsonUI) PrintBlock(block string) {
	ui.uiResp.Blocks = append(ui.uiResp.Blocks, block)
}

func (ui *jsonUI) PrintErrorBlock(block string) {
	ui.uiResp.Blocks = append(ui.uiResp.Blocks, block)
}

func (ui *jsonUI) PrintTable(table Table) {
	table.FillFirstColumn = true

	var header []string

	if len(table.HeaderVals) > 0 {
		for _, val := range table.HeaderVals {
			header = append(header, val.String())
		}
	} else if len(table.Header) > 0 {
		header = table.Header
	}

	resp := tableResp{
		Content: table.Content,
		Header:  header,
		Rows:    ui.stringRows(table.AsRows()),
		Notes:   table.Notes,
	}

	ui.uiResp.Tables = append(ui.uiResp.Tables, resp)
}

func (ui *jsonUI) AskForText(_ string) (string, error) {
	panic("Cannot ask for input in JSON UI")
}

func (ui *jsonUI) AskForChoice(_ string, _ []string) (int, error) {
	panic("Cannot ask for a choice in JSON UI")
}

func (ui *jsonUI) AskForPassword(_ string) (string, error) {
	panic("Cannot ask for password in JSON UI")
}

func (ui *jsonUI) AskForConfirmation() error {
	panic("Cannot ask for confirmation in JSON UI")
}

func (ui *jsonUI) IsInteractive() bool {
	return ui.parent.IsInteractive()
}

func (ui *jsonUI) Flush() {
	defer ui.parent.Flush()

	if !reflect.DeepEqual(ui.uiResp, uiResp{}) {
		bytes, err := json.MarshalIndent(ui.uiResp, "", "    ")
		if err != nil {
			ui.logger.Error(ui.logTag, "Failed to marshal UI response")
			return
		}

		ui.parent.PrintBlock(string(bytes))
	}
}

func (ui *jsonUI) stringRows(vals [][]Value) [][]string {
	var result [][]string

	for _, row := range vals {
		var strs []string

		for _, v := range row {
			strs = append(strs, v.String())
		}

		result = append(result, strs)
	}

	return result
}

func (ui *jsonUI) addLine(pattern string, args []interface{}) {
	msg := fmt.Sprintf(pattern, args...)
	ui.uiResp.Lines = append(ui.uiResp.Lines, msg)
	ui.logger.Debug(ui.logTag, msg)
}
