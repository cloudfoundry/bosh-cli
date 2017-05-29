package cmd

import (
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type Diff struct {
	Lines [][]interface{}
}

func NewDiff(lines [][]interface{}) Diff {
	return Diff{
		Lines: lines,
	}
}

func (d Diff) Print(ui boshui.UI) {
	for _, line := range d.Lines {
		lineMod, _ := line[1].(string)

		if lineMod == "added" {
			ui.BeginLinef("+ %s\n", line[0])
		} else if lineMod == "removed" {
			ui.BeginLinef("- %s\n", line[0])
		} else {
			ui.BeginLinef("  %s\n", line[0])
		}
	}
}
