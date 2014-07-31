package main

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	"os"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
)

const mainLogTag = "main"

func main() {
	logger := boshlog.NewLogger(boshlog.LevelError)
	defer logger.HandlePanic("Main")

	cmdFactory := bmcmd.NewFactory(logger)
	cmdRunner := bmcmd.NewRunner(cmdFactory)

	err := cmdRunner.Run(os.Args[1:])
	if err != nil {
		logger.Error(mainLogTag, "BOSH Micro CLI failed with: '%s'", err.Error())
		os.Exit(1)
	}
}
