package main

import (
	"os"
	"path"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshlogfile "github.com/cloudfoundry/bosh-agent/logger/file"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

const mainLogTag = "main"

func main() {
	logger := newLogger()
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

func newLogger() boshlog.Logger {
	logLevelString := os.Getenv("BOSH_MICRO_LOG_LEVEL")
	level := boshlog.LevelError
	if logLevelString != "" {
		var err error
		level, err = boshlog.Levelify(logLevelString)
		if err != nil {
			err = bosherr.WrapError(err, "Invalid BOSH_MICRO_LOG_LEVEL value")
			logger := boshlog.NewLogger(boshlog.LevelError)
			fail(err, logger)
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
		fail(err, logger)
	}
	return logger
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
	logger.Error(mainLogTag, "BOSH Micro CLI failed: %s", err.Error())
	os.Exit(1)
}
