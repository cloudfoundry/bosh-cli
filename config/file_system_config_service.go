package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

const (
	tagString = "Config"
)

type fileSystemConfigService struct {
	configPath string
	fs         boshsys.FileSystem
	logger     boshlog.Logger
}

func NewFileSystemConfigService(logger boshlog.Logger, fs boshsys.FileSystem, configPath string) Service {
	return fileSystemConfigService{
		configPath: configPath,
		fs:         fs,
		logger:     logger,
	}
}

type DeploymentFile struct {
	UUID string
}

func (s fileSystemConfigService) Load() (Config, error) {
	content, err := s.fs.ReadFile(s.configPath)
	if err != nil {
		s.logger.Info("No config file found at %s, using default configuration", s.configPath)
		return Config{}, nil
	}

	config := Config{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return Config{}, bosherr.WrapError(err, "Unmarshalling JSON config file `%s'", s.configPath)
	}

	deploymentFileContents, err := s.fs.ReadFile(config.DeploymentFile())
	if err != nil {
		return Config{}, bosherr.WrapError(err, "Loading deployment file `%s'", config.DeploymentFile())

	}
	s.logger.Debug(tagString, "Deployment File Contents %#s", deploymentFileContents)

	deploymentFile := DeploymentFile{}

	err = json.Unmarshal(deploymentFileContents, &deploymentFile)
	if err != nil {
		return Config{}, bosherr.WrapError(err, "Unmarshalling deployment file `%s'", config.DeploymentFile())
	}
	config.DeploymentUUID = deploymentFile.UUID
	return config, nil
}

func (s fileSystemConfigService) Save(config Config) error {
	jsonContent, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling config into JSON")
	}

	err = s.fs.WriteFile(s.configPath, jsonContent)
	if err != nil {
		return bosherr.WrapError(err, "Writing config file `%s'", s.configPath)
	}

	jsonContent, err = json.MarshalIndent(DeploymentFile{UUID: config.DeploymentUUID}, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling config into JSON")
	}

	err = s.fs.WriteFile(config.DeploymentFile(), jsonContent)
	if err != nil {
		return bosherr.WrapError(err, "Writing deployment file `%s'", s.configPath)
	}

	return nil
}
