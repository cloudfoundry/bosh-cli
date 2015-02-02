package manifest

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type NetworkType string

func (n NetworkType) String() string {
	return string(n)
}

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

func (n Network) CloudProperties() (bmproperty.Map, error) {
	return bmproperty.BuildMap(n.RawCloudProperties)
}

// Interface returns a property map representing a generic network interface.
// Expected Keys: ip, type, cloud properties.
// Optional Keys: netmask, gateway, dns
func (n Network) Interface() (bmproperty.Map, error) {
	cloudProperties, err := n.CloudProperties()
	if err != nil {
		return bmproperty.Map{}, err
	}

	iface := bmproperty.Map{
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
