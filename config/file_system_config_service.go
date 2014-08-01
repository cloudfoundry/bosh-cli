package config

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
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

func (s fileSystemConfigService) Load() (Config, error) {
	content, err := s.fs.ReadFile(s.configPath)
	if err != nil {
		s.logger.Info("No config file found at %s, using default configuration", s.configPath)
		return Config{}, nil
	}

	config := Config{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return Config{}, bosherr.WrapError(err, "Unmarshalling JSON config file '%s'", s.configPath)
	}

	return config, nil
}

func (s fileSystemConfigService) Save(config Config) error {
	jsonContent, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling config into JSON")
	}

	err = s.fs.WriteFile(s.configPath, jsonContent)
	if err != nil {
		return bosherr.WrapError(err, "Writing config file '%s'", s.configPath)
	}

	return nil
}
