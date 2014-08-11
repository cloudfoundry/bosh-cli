package main

import (
	"os"
	"path"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

const mainLogTag = "main"

func main() {
	logger := boshlog.NewLogger(boshlog.LevelError)
	defer logger.HandlePanic("Main")
	fileSystem := boshsys.NewOsFileSystem(logger)
	config, configService := loadConfig(logger, fileSystem)
	ui := bmui.NewDefaultUI(os.Stdout, os.Stderr)
	runner := boshsys.NewExecCmdRunner(logger)
	extractor := bmtar.NewCmdExtractor(runner, logger)
	releaseValidator := bmrelease.NewValidator(fileSystem)
	cpiReleaseValidator := bmrelease.NewCpiValidator()

	cmdFactory := bmcmd.NewFactory(config, configService, fileSystem, ui, extractor, releaseValidator, cpiReleaseValidator)
	cmdRunner := bmcmd.NewRunner(cmdFactory)

	err := cmdRunner.Run(os.Args[1:])
	if err != nil {
		fail(err, logger)
	}
}

func loadConfig(logger boshlog.Logger, fileSystem boshsys.FileSystem) (bmconfig.Config, bmconfig.Service) {
	configPath := os.Getenv("HOME")
	configService := bmconfig.NewFileSystemConfigService(logger, fileSystem, path.Join(configPath, ".bosh_micro.json"))
	config, err := configService.Load()
	if err != nil {
		fail(err, logger)
	}
	return config, configService
}

func fail(err error, logger boshlog.Logger) {
	logger.Error(mainLogTag, "BOSH Micro CLI failed with: `%s'", err.Error())
	os.Exit(1)
}
