package main

import (
	"os"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

const mainLogTag = "main"

func main() {
	logger := boshlog.NewLogger(boshlog.LevelError)
	defer logger.HandlePanic("Main")

	fileSystem := boshsys.NewOsFileSystem(logger)
	ui := bmui.NewDefaultUI(os.Stdout, os.Stderr)
	boshMicroPath := os.Getenv("HOME")
	cmdFactory := bmcmd.NewFactory(fileSystem, ui, boshMicroPath)

	cmdRunner := bmcmd.NewRunner(cmdFactory)

	err := cmdRunner.Run(os.Args[1:])
	if err != nil {
		logger.Error(mainLogTag, "BOSH Micro CLI failed with: '%s'", err.Error())
		os.Exit(1)
	}
}
