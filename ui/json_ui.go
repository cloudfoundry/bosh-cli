package ui

import (
	"encoding/json"
	"fmt"
	"reflect"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
	"strconv"
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
	Header  map[string]string
	Rows    []map[string]string
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

	header := map[string]string{}
	var rawHeaders []Header

	if len(table.Header) > 0 {
		for i, val := range table.Header {
			if !val.Visible {
				continue
			}

			if val.Key == string(UNKNOWN_HEADER_MAPPING) {
				val.Key = strconv.Itoa(i)
			}

			header[val.Key] = val.Title
			rawHeaders = append(rawHeaders, val)
		}
	} else if len(table.AsRows()) > 0 {
		for i, _ := range table.AsRows()[0] {
			val := Header{
				Key:     fmt.Sprintf("col_%d", i),
				Visible: true,
			}
			header[val.Key] = val.Title
			rawHeaders = append(rawHeaders, val)
		}
	}

	resp := tableResp{
		Content: table.Content,
		Header:  header,
		Rows:    ui.stringRows(rawHeaders, table.AsRows()),
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

func (ui *jsonUI) stringRows(header []Header, vals [][]Value) []map[string]string {
	result := []map[string]string{}

	for _, row := range vals {
		strs := map[string]string{}

		for i, v := range row {
			if !header[i].Visible {
				continue
			}

			strs[header[i].Key] = v.String()
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
