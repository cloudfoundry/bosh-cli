package cmd

import (
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

func printManifestDiff(ui boshui.UI, diff [][]interface{}) {
	for _, line := range diff {
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
