package manifest

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type NetworkType string

func (n NetworkType) String() string {
	return string(n)
}

type NetworkInterface map[string]interface{}

const (
	Dynamic NetworkType = "dynamic"
	Manual  NetworkType = "manual"
	VIP     NetworkType = "vip"
)

type Network struct {
	Name               string                      `yaml:"name"`
	Type               NetworkType                 `yaml:"type"`
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
	IP                 string                      `yaml:"ip"`
	Netmask            string                      `yaml:"netmask"`
	Gateway            string                      `yaml:"gateway"`
	DNS                []string                    `yaml:"dns"`
}

func (n Network) CloudProperties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(n.RawCloudProperties)
}

func (n Network) Interface() (NetworkInterface, error) {
	cloudProperties, err := n.CloudProperties()
	if err != nil {
		return NetworkInterface{}, err
	}

	iface := NetworkInterface{
		"type":             n.Type.String(),
		"ip":               n.IP,
		"cloud_properties": cloudProperties,
	}

	if n.Netmask != "" {
		iface["netmask"] = n.Netmask
	}

	if n.Gateway != "" {
		iface["gateway"] = n.Gateway
	}

	if len(n.DNS) > 0 {
		iface["dns"] = n.DNS
	}

	return iface, nil
}
