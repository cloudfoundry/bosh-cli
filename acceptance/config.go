package acceptance

import (
	"encoding/json"
	"errors"
	"os"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Config struct {
	VMUsername      string `json:"vm_username"`
	VMIP            string `json:"vm_ip"`
	VMPort          string `json:"vm_port"`
	PrivateKeyPath  string `json:"private_key_path"`
	StemcellURL     string `json:"stemcell_url"`
	StemcellPath    string `json:"stemcell_path"`
	CpiReleaseURL   string `json:"cpi_release_url"`
	CpiReleasePath  string `json:"cpi_release_path"`
	DummyReleasePath string `json:"dummy_release_path"`
}

func NewConfig(fs boshsys.FileSystem) (*Config, error) {
	path := os.Getenv("BOSH_MICRO_CONFIG_PATH")
	if path == "" {
		return &Config{}, errors.New("Must provide config file via BOSH_MICRO_CONFIG_PATH environment variable")
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
	if c.VMUsername == "" {
		return errors.New("Must provide 'vm_username' in config")
	}

	if c.VMIP == "" {
		return errors.New("Must provide 'vm_ip' in config")
	}

	if c.PrivateKeyPath == "" {
		return errors.New("Must provide 'private_key_path' in config")
	}

	if c.StemcellURL == "" && c.StemcellPath == "" {
		return errors.New("Must provide 'stemcell_url' or 'stemcell_path' in config")
	}

	if c.CpiReleaseURL == "" && c.CpiReleasePath == "" {
		return errors.New("Must provide 'cpi_release_url' or 'cpi_release_path' in config")
	}

	return nil
}
