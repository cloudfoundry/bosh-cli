package ui

import (
	"encoding/json"
	"fmt"
	"reflect"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
	"strings"
	"unicode"
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
	var headerVals []string

	if len(table.HeaderVals) > 0 {
		for _, val := range table.HeaderVals {
			header[keyifyHeader(val.String())] = val.String()
			headerVals = append(headerVals, val.String())
		}
	} else if len(table.Header) > 0 {
		for _, val := range table.Header {
			header[keyifyHeader(val)] = val
			headerVals = append(headerVals, val)
		}
	} else if len(table.AsRows()) > 0 {
		for i, _ := range table.AsRows()[0] {
			s := fmt.Sprintf("col_%d", i)
			header[s] = ""
			headerVals = append(headerVals, s)
		}
	}

	resp := tableResp{
		Content: table.Content,
		Header:  header,
		Rows:    ui.stringRows(headerVals, table.AsRows()),
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

func (ui *jsonUI) stringRows(header []string, vals [][]Value) []map[string]string {
	result := []map[string]string{}

	for _, row := range vals {
		strs := map[string]string{}

		for i, v := range row {
			strs[keyifyHeader(header[i])] = v.String()
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

func keyifyHeader(header string) string {
	splittedStrings := strings.Split(cleanHeader(header), " ")
	splittedTrimmedStrings := []string{}
	for _, s := range splittedStrings {
		if s != "" {
			splittedTrimmedStrings = append(splittedTrimmedStrings, strings.Trim(s, " "))
		}
	}

	return strings.Join(splittedTrimmedStrings, "_")
}

func cleanHeader(header string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		} else {
			return ' '
		}
	}, header)
}
