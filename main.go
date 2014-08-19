package main

import (
	"os"
	"path"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bmrelcomp "github.com/cloudfoundry/bosh-micro-cli/release/compile"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

const mainLogTag = "main"

func main() {
	logger := boshlog.NewLogger(boshlog.LevelError)
	defer logger.HandlePanic("Main")
	fileSystem := boshsys.NewOsFileSystem(logger)

	boshMicroDir := path.Join(os.Getenv("HOME"), ".bosh_micro")
	fileSystem.MkdirAll(boshMicroDir, os.ModePerm)

	config, configService := loadConfig(logger, fileSystem)

	runner := boshsys.NewExecCmdRunner(logger)
	extractor := bmtar.NewCmdExtractor(runner, logger)

	ui := bmui.NewDefaultUI(os.Stdout, os.Stderr)
	boshValidator := bmrelvalidation.NewBoshValidator(fileSystem)
	cpiReleaseValidator := bmrelvalidation.NewCpiValidator()
	releaseValidator := bmrelvalidation.NewValidator(boshValidator, cpiReleaseValidator, ui)

	compressor := boshcmd.NewTarballCompressor(runner, fileSystem)
	uuidGenerator := boshuuid.NewGenerator()
	blobDir := path.Join(boshMicroDir, "blobs")
	fileSystem.MkdirAll(blobDir, os.ModePerm)
	options := map[string]interface{}{"blobstore_path": blobDir}

	blobstore := boshblob.NewSHA1VerifiableBlobstore(
		boshblob.NewLocalBlobstore(fileSystem, uuidGenerator, options),
	)

	indexFilePath := path.Join(boshMicroDir, "index.json")
	index := bmindex.NewFileIndex(indexFilePath, fileSystem)
	compiledPackageRepo := bmrelcomp.NewCompiledPackageRepo(index)
	packageCompiler := bmrelcomp.NewPackageCompiler(
		runner,
		path.Join(boshMicroDir, "packages"),
		fileSystem,
		compressor,
		blobstore,
		compiledPackageRepo,
	)
	da := bmrelcomp.NewDependencyAnalysis()
	releaseCompiler := bmrelcomp.NewReleaseCompiler(da, packageCompiler)

	cmdFactory := bmcmd.NewFactory(
		config,
		configService,
		fileSystem,
		ui,
		extractor,
		releaseValidator,
		releaseCompiler,
	)
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
