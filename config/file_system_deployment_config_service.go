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

type StemcellRecord struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	CID     string `json:"cid"`
}

type DeploymentFile struct {
	UUID              string           `json:"uuid"`
	CurrentVMCID      string           `json:"current_vm_cid"`
	CurrentStemcellID string           `json:"current_stemcell_id"`
	CurrentDiskID     string           `json:"current_disk_id"`
	CurrentReleaseID  string           `json:"current_release_id"`
	Disks             []DiskRecord     `json:"disks"`
	Stemcells         []StemcellRecord `json:"stemcells"`
}

func (s fileSystemDeploymentConfigService) Load() (DeploymentConfig, error) {
	config := DeploymentConfig{}

	if !s.fs.FileExists(s.configPath) {
		return config, nil
	}

	deploymentFileContents, err := s.fs.ReadFile(s.configPath)
	if err != nil {
		return config, bosherr.WrapError(err, "Reading deployment config file `%s'", s.configPath)
	}
	s.logger.Debug(s.logTag, "Deployment File Contents %#s", deploymentFileContents)

	deploymentFile := DeploymentFile{}

	err = json.Unmarshal(deploymentFileContents, &deploymentFile)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshalling deployment config file `%s'", s.configPath)
	}

	config.DeploymentUUID = deploymentFile.UUID
	config.CurrentVMCID = deploymentFile.CurrentVMCID
	config.CurrentDiskID = deploymentFile.CurrentDiskID
	config.Disks = deploymentFile.Disks
	config.Stemcells = deploymentFile.Stemcells

	return config, nil
}

func (s fileSystemDeploymentConfigService) Save(config DeploymentConfig) error {
	deploymentFile := DeploymentFile{
		UUID:          config.DeploymentUUID,
		CurrentVMCID:  config.CurrentVMCID,
		CurrentDiskID: config.CurrentDiskID,
		Stemcells:     config.Stemcells,
		Disks:         config.Disks,
	}
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
