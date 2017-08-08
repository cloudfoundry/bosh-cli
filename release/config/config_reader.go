package config

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	yaml "gopkg.in/yaml.v2"
)

type ConfigReaderImpl struct {
	fs boshsys.FileSystem
}

func NewConfigReaderImpl(fs boshsys.FileSystem) *ConfigReaderImpl {
	return &ConfigReaderImpl{fs}
}

func (r ConfigReaderImpl) Read(path string) (*Config, error) {
	bytes, err := r.fs.ReadFile(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading config '%s'", path)
	}
	publicSchema := Config{}
	err = yaml.Unmarshal(bytes, &publicSchema)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Unmarshalling config '%s'", path)
	}
	return &publicSchema, nil
}
