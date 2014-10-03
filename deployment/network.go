package deployment

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type NetworkType string

func (n NetworkType) String() string {
	return string(n)
}

const (
	Dynamic NetworkType = "dynamic"
)

type Network struct {
	Name               string                      `yaml:"name"`
	Type               NetworkType                 `yaml:"type"`
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

func (n Network) CloudProperties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(n.RawCloudProperties)
}

func (n Network) Spec() (map[string]interface{}, error) {
	cloudProperties, err := n.CloudProperties()
	if err != nil {
		return map[string]interface{}{}, err
	}

	spec := map[string]interface{}{
		n.Name: map[string]interface{}{
			"type":             n.Type.String(),
			"cloud_properties": cloudProperties,
		},
	}

	return spec, nil
}
