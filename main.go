package main

import (
	"os"
	"path"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

const mainLogTag = "main"

func main() {
	debugLogEnabled := os.Getenv("BOSH_MICRO_LOG")
	var logger boshlog.Logger
	if debugLogEnabled != "" {
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr, os.Stderr)
	} else {
		logger = boshlog.NewLogger(boshlog.LevelError)
	}
	defer logger.HandlePanic("Main")
	fileSystem := boshsys.NewOsFileSystem(logger)
	workspace := path.Join(os.Getenv("HOME"), ".bosh_micro")
	userConfigPath := path.Join(os.Getenv("HOME"), ".bosh_micro.json")
	config, configService := loadUserConfig(userConfigPath, fileSystem, logger)

	uuidGenerator := boshuuid.NewGenerator()

	ui := bmui.NewUI(os.Stdout, os.Stderr, logger)

	cmdFactory := bmcmd.NewFactory(
		config,
		configService,
		fileSystem,
		ui,
		logger,
		uuidGenerator,
		workspace,
	)

	cmdRunner := bmcmd.NewRunner(cmdFactory)

	err := cmdRunner.Run(os.Args[1:])
	if err != nil {
		fail(err, logger)
	}
}

func loadUserConfig(userConfigPath string, fileSystem boshsys.FileSystem, logger boshlog.Logger) (
	bmconfig.UserConfig,
	bmconfig.UserConfigService,
) {
	userConfigService := bmconfig.NewFileSystemUserConfigService(userConfigPath, fileSystem, logger)
	userConfig, err := userConfigService.Load()
	if err != nil {
		fail(err, logger)
	}

	return userConfig, userConfigService
}

func fail(err error, logger boshlog.Logger) {
	logger.Error(mainLogTag, "BOSH Micro CLI failed with: %s", err.Error())
	os.Exit(1)
}
