package manifest

import (
	"encoding/hex"
	"fmt"
	"net"

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
	DNS             []string
	Subnets         []Subnet
}

type Subnet struct {
	Range           string
	Gateway         string
	DNS             []string
	CloudProperties biproperty.Map
}

// Interface returns a property map representing a generic network interface.
// Expected Keys: ip, type, cloud properties.
// Optional Keys: netmask, gateway, dns
func (n Network) Interface(staticIPs []string) biproperty.Map {
	networkInterface := biproperty.Map{
		"type": n.Type.String(),
	}

	if n.Type == Manual {
		networkInterface["gateway"] = n.Subnets[0].Gateway
		if len(n.Subnets[0].DNS) > 0 {
			networkInterface["dns"] = n.Subnets[0].DNS
		}

		_, ipNet, _ := net.ParseCIDR(n.Subnets[0].Range)
		a, _ := hex.DecodeString(ipNet.Mask.String())
		networkInterface["netmask"] = fmt.Sprintf("%v.%v.%v.%v", a[0], a[1], a[2], a[3])

		networkInterface["cloud_properties"] = n.Subnets[0].CloudProperties
	} else {
		networkInterface["cloud_properties"] = n.CloudProperties
	}

	if n.Type == Dynamic && len(n.DNS) > 0 {
		networkInterface["dns"] = n.DNS
	}

	if len(staticIPs) > 0 {
		networkInterface["ip"] = staticIPs[0]
	}

	return networkInterface
}
