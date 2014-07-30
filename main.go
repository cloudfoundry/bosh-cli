package main

import (
	"fmt"
	"os"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
)

func main() {
	cmdFactory := bmcmd.NewFactory()
	cmdRunner := bmcmd.NewRunner(cmdFactory)

	err := cmdRunner.Run(os.Args[1:])
	if err != nil {
		fmt.Println(fmt.Sprintf("Error running bosh-micro-cli - %s", err.Error()))

		os.Exit(1)
	}
}
