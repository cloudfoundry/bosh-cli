package acceptance

import (
	"encoding/json"
	"errors"
	"os"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Config struct {
	StemcellPath             string `json:"stemcell_path"`
	CPIReleasePath           string `json:"cpi_release_path"`
	DummyCompiledReleasePath string `json:"dummy_compiled_release_path"`
}

func NewConfig(fs boshsys.FileSystem) (*Config, error) {
	path := os.Getenv("BOSH_INIT_CONFIG_PATH")
	if path == "" {
		return &Config{}, errors.New("Must provide config file via BOSH_INIT_CONFIG_PATH environment variable")
	}

	configContents, err := fs.ReadFile(path)
	if err != nil {
		return &Config{}, err
	}

	var config Config
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return &Config{}, err
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.StemcellPath == "" {
		return errors.New("Must provide 'stemcell_path' in config")
	}

	if c.CPIReleasePath == "" {
		return errors.New("Must provide 'cpi_release_path' in config")
	}

	if c.DummyCompiledReleasePath == "" {
		return errors.New("Must provide 'dummy_compiled_release_path' in config")
	}

	return nil
}
