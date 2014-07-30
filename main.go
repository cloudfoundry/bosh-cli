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
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
