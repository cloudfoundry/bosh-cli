package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
	"path"
)

type fileSystemDeploymentStateService struct {
	configPath    string
	fs            boshsys.FileSystem
	uuidGenerator boshuuid.Generator
	logger        boshlog.Logger
	logTag        string
}

func NewFileSystemDeploymentStateService(fs boshsys.FileSystem, uuidGenerator boshuuid.Generator, logger boshlog.Logger, deploymentStatePath string) DeploymentStateService {
	return &fileSystemDeploymentStateService{
		configPath:    deploymentStatePath,
		fs:            fs,
		uuidGenerator: uuidGenerator,
		logger:        logger,
		logTag:        "config",
	}
}

func DeploymentStatePath(deploymentManifestPath string) string {
	return path.Join(path.Dir(deploymentManifestPath), "deployment.json")
}

func (s *fileSystemDeploymentStateService) Path() string {
	return s.configPath
}

func (s *fileSystemDeploymentStateService) Exists() bool {
	return s.fs.FileExists(s.configPath)
}

func (s *fileSystemDeploymentStateService) Load() (DeploymentState, error) {
	if s.configPath == "" {
		panic("configPath not yet set!")
	}

	s.logger.Debug(s.logTag, "Loading deployment config: %s", s.configPath)

	deploymentState := &DeploymentState{}

	if s.fs.FileExists(s.configPath) {
		deploymentStateFileContents, err := s.fs.ReadFile(s.configPath)
		if err != nil {
			return DeploymentState{}, bosherr.WrapErrorf(err, "Reading deployment config file '%s'", s.configPath)
		}
		s.logger.Debug(s.logTag, "Deployment File Contents %#s", deploymentStateFileContents)

		err = json.Unmarshal(deploymentStateFileContents, deploymentState)
		if err != nil {
			return DeploymentState{}, bosherr.WrapErrorf(err, "Unmarshalling deployment config file '%s'", s.configPath)
		}
	}

	err := s.initDefaults(deploymentState)
	if err != nil {
		return DeploymentState{}, bosherr.WrapErrorf(err, "Initializing deployment config defaults")
	}

	return *deploymentState, nil
}

func (s *fileSystemDeploymentStateService) Save(deploymentState DeploymentState) error {
	if s.configPath == "" {
		panic("configPath not yet set!")
	}

	s.logger.Debug(s.logTag, "Saving deployment config %#v", deploymentState)

	jsonContent, err := json.MarshalIndent(deploymentState, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling deployment config into JSON")
	}

	err = s.fs.WriteFile(s.configPath, jsonContent)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing deployment config file '%s'", s.configPath)
	}

	return nil
}

func (s *fileSystemDeploymentStateService) initDefaults(deploymentState *DeploymentState) error {
	if deploymentState.DirectorID == "" {
		uuid, err := s.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "Generating DirectorID")
		}
		deploymentState.DirectorID = uuid

		err = s.Save(*deploymentState)
		if err != nil {
			return bosherr.WrapError(err, "Saving deployment config")
		}
	}

	return nil
}
