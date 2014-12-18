package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

func NewFileSystemUserConfigService(configPath string, fs boshsys.FileSystem, logger boshlog.Logger) UserConfigService {
	return &fileSystemUserConfigService{
		configPath: configPath,
		fs:         fs,
		logger:     logger,
	}
}

type fileSystemUserConfigService struct {
	configPath string
	fs         boshsys.FileSystem
	logger     boshlog.Logger
}

func (s *fileSystemUserConfigService) Load() (UserConfig, error) {
	config := UserConfig{}

	configContents, err := s.fs.ReadFile(s.configPath)
	if err != nil {
		return config, nil
	}

	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return config, bosherr.WrapErrorf(err, "Unmarshal config '%s'", configContents)
	}

	return config, nil
}

func (s *fileSystemUserConfigService) Save(config UserConfig) error {
	jsonContent, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling user config into JSON")
	}

	err = s.fs.WriteFile(s.configPath, jsonContent)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing user config file '%s'", s.configPath)
	}

	return nil
}
