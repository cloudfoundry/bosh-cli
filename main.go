package main

import (
	"os"
	"path"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshlogfile "github.com/cloudfoundry/bosh-agent/logger/file"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bicmd "github.com/cloudfoundry/bosh-init/cmd"
	biconfig "github.com/cloudfoundry/bosh-init/config"

	biui "github.com/cloudfoundry/bosh-init/ui"
	biuifmt "github.com/cloudfoundry/bosh-init/ui/fmt"
)

const mainLogTag = "main"

func main() {
	logger := newLogger()
	defer logger.HandlePanic("Main")
	fileSystem := boshsys.NewOsFileSystem(logger)
	workspaceRootPath := path.Join(os.Getenv("HOME"), ".bosh_micro")
	userConfigPath := path.Join(os.Getenv("HOME"), ".bosh_micro.json")
	ui := biui.NewConsoleUI(logger)
	config, configService := loadUserConfig(userConfigPath, fileSystem, ui, logger)

	uuidGenerator := boshuuid.NewGenerator()

	timeService := boshtime.NewConcreteService()

	cmdFactory := bicmd.NewFactory(
		config,
		configService,
		fileSystem,
		ui,
		timeService,
		logger,
		uuidGenerator,
		workspaceRootPath,
	)

	cmdRunner := bicmd.NewRunner(cmdFactory)
	stage := biui.NewStage(ui, timeService, logger)
	err := cmdRunner.Run(stage, os.Args[1:]...)
	if err != nil {
		fail(err, ui, logger)
	}
}

func newLogger() boshlog.Logger {
	logLevelString := os.Getenv("BOSH_MICRO_LOG_LEVEL")
	level := boshlog.LevelNone
	if logLevelString != "" {
		var err error
		level, err = boshlog.Levelify(logLevelString)
		if err != nil {
			err = bosherr.WrapError(err, "Invalid BOSH_MICRO_LOG_LEVEL value")
			logger := boshlog.NewLogger(boshlog.LevelError)
			ui := biui.NewConsoleUI(logger)
			fail(err, ui, logger)
		}
	}

	logPath := os.Getenv("BOSH_MICRO_LOG_PATH")
	if logPath != "" {
		return newFileLogger(logPath, level)
	}

	return boshlog.NewLogger(level)
}

func newFileLogger(logPath string, level boshlog.LogLevel) boshlog.Logger {
	// Log file logger errors to the STDERR logger
	logger := boshlog.NewLogger(boshlog.LevelError)
	fileSystem := boshsys.NewOsFileSystem(logger)

	// log file will be closed by process exit
	// log file readable by all
	logger, _, err := boshlogfile.New(level, logPath, boshlogfile.DefaultLogFileMode, fileSystem)
	if err != nil {
		logger := boshlog.NewLogger(boshlog.LevelError)
		ui := biui.NewConsoleUI(logger)
		fail(err, ui, logger)
	}
	return logger
}

func loadUserConfig(userConfigPath string, fileSystem boshsys.FileSystem, ui biui.UI, logger boshlog.Logger) (
	biconfig.UserConfig,
	biconfig.UserConfigService,
) {
	userConfigService := biconfig.NewFileSystemUserConfigService(userConfigPath, fileSystem, logger)
	userConfig, err := userConfigService.Load()
	if err != nil {
		fail(err, ui, logger)
	}

	return userConfig, userConfigService
}

func fail(err error, ui biui.UI, logger boshlog.Logger) {
	logger.Error(mainLogTag, err.Error())
	ui.ErrorLinef("")
	ui.ErrorLinef(biuifmt.MultilineError(err))
	os.Exit(1)
}
