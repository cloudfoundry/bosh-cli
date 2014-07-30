package main

import (
	"os"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
)

func main() {
	cmdFactory := bmcmd.NewFactory()
	cmdRunner := bmcmd.NewRunner(cmdFactory)

	err := cmdRunner.Run(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
