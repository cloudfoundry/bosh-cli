package stemcell

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type Stemcell struct {
	Manifest Manifest
}

type Manifest struct {
	ImagePath          string
	Name               string
	Version            string
	SHA1               string
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

func (m Manifest) CloudProperties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(m.RawCloudProperties)
}
