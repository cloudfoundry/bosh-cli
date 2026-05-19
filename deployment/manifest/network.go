package manifest

import (
	"fmt"
	"net"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
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
	AZs             []string
}

// ContainsAZ reports whether this subnet is associated with the given AZ name.
// A subnet with no AZ labels is considered to match any AZ (backward compatible).
func (s Subnet) ContainsAZ(az string) bool {
	if len(s.AZs) == 0 {
		return true
	}
	for _, a := range s.AZs {
		if a == az {
			return true
		}
	}
	return false
}

// Interface returns a property map representing a generic network interface.
// Expected Keys: ip, type, cloud properties.
// Optional Keys: netmask, gateway, dns
//
// Deprecated: prefer InterfaceForAZ which selects the correct subnet when AZs
// are configured.
func (n Network) Interface(staticIPs []string, networkDefaults []NetworkDefault) (biproperty.Map, error) {
	return n.InterfaceForAZ(staticIPs, networkDefaults, "")
}

// InterfaceForAZ returns a property map for the network interface to be used by
// an instance in the given AZ. When az is non-empty and the network has subnets
// with AZ labels, the first matching subnet is selected; otherwise the first
// subnet is used (backward-compatible behaviour).
func (n Network) InterfaceForAZ(staticIPs []string, networkDefaults []NetworkDefault, az string) (biproperty.Map, error) {
	networkInterface := biproperty.Map{
		"type": n.Type.String(),
	}

	if n.Type == Manual {
		subnet := n.subnetForAZ(az)
		networkInterface["gateway"] = subnet.Gateway
		if len(subnet.DNS) > 0 {
			networkInterface["dns"] = subnet.DNS
		}

		_, ipNet, err := net.ParseCIDR(subnet.Range)
		if err != nil {
			return biproperty.Map{}, bosherr.WrapError(err, "Failed to parse subnet range")
		}

		networkInterface["netmask"] = ipMaskString(ipNet.Mask)
		networkInterface["cloud_properties"] = subnet.CloudProperties
	} else {
		networkInterface["cloud_properties"] = n.CloudProperties
	}

	if n.Type == Dynamic && len(n.DNS) > 0 {
		networkInterface["dns"] = n.DNS
	}

	if len(staticIPs) > 0 {
		networkInterface["ip"] = staticIPs[0]
	}

	if len(networkDefaults) > 0 {
		networkInterface["default"] = networkDefaults
	}

	return networkInterface, nil
}

// subnetForAZ returns the first subnet whose AZ set contains az.
// If no labelled subnet matches (or az is empty), the first subnet is returned.
func (n Network) subnetForAZ(az string) Subnet {
	if az != "" {
		for _, s := range n.Subnets {
			if len(s.AZs) > 0 && s.ContainsAZ(az) {
				return s
			}
		}
	}
	return n.Subnets[0]
}

func ipMaskString(ipMask net.IPMask) string {
	ip := net.IP(ipMask)

	if p4 := ip.To4(); len(p4) == net.IPv4len {
		return ip.String()
	}

	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
		[]byte(ip[0:2]), []byte(ip[2:4]), []byte(ip[4:6]), []byte(ip[6:8]),
		[]byte(ip[8:10]), []byte(ip[10:12]), []byte(ip[12:14]), []byte(ip[14:16]))
}
