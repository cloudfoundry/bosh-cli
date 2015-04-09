package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
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
	Name            string
	Type            NetworkType
	CloudProperties biproperty.Map
	IP              string
	Netmask         string
	Gateway         string
	DNS             []string
}

// Interface returns a property map representing a generic network interface.
// Expected Keys: ip, type, cloud properties.
// Optional Keys: netmask, gateway, dns
func (n Network) Interface() biproperty.Map {
	iface := biproperty.Map{
		"type":             n.Type.String(),
		"ip":               n.IP,
		"cloud_properties": n.CloudProperties,
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

	return iface
}
