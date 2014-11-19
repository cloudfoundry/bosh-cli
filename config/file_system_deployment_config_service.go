package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type fileSystemDeploymentConfigService struct {
	configPath string
	fs         boshsys.FileSystem
	logger     boshlog.Logger
	logTag     string
}

func NewFileSystemDeploymentConfigService(configPath string, fs boshsys.FileSystem, logger boshlog.Logger) DeploymentConfigService {
	return fileSystemDeploymentConfigService{
		configPath: configPath,
		fs:         fs,
		logger:     logger,
		logTag:     "config",
	}
}

func (s fileSystemDeploymentConfigService) Load() (DeploymentFile, error) {
	if !s.fs.FileExists(s.configPath) {
		return DeploymentFile{}, nil
	}

	deploymentFileContents, err := s.fs.ReadFile(s.configPath)
	if err != nil {
		return DeploymentFile{}, bosherr.WrapError(err, "Reading deployment config file `%s'", s.configPath)
	}
	s.logger.Debug(s.logTag, "Deployment File Contents %#s", deploymentFileContents)

	deploymentFile := DeploymentFile{}

	err = json.Unmarshal(deploymentFileContents, &deploymentFile)
	if err != nil {
		return DeploymentFile{}, bosherr.WrapError(err, "Unmarshalling deployment config file `%s'", s.configPath)
	}

	return deploymentFile, nil
}

func (s fileSystemDeploymentConfigService) Save(deploymentFile DeploymentFile) error {
	jsonContent, err := json.MarshalIndent(deploymentFile, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling deployment config into JSON")
	}

	err = s.fs.WriteFile(s.configPath, jsonContent)
	if err != nil {
		return bosherr.WrapError(err, "Writing deployment config file `%s'", s.configPath)
	}

	return nil
}
