package settings

import (
	"fmt"
)

const (
	RootUsername        = "root"
	VCAPUsername        = "vcap"
	AdminGroup          = "admin"
	EphemeralUserPrefix = "bosh_"
)

type Settings struct {
	AgentID      string    `json:"agent_id"`
	Blobstore    Blobstore `json:"blobstore"`
	Disks        Disks     `json:"disks"`
	Env          Env       `json:"env"`
	Networks     Networks  `json:"networks"`
	Ntp          []string  `json:"ntp"`
	Mbus         string    `json:"mbus"`
	VM           VM        `json:"vm"`
	TrustedCerts string    `json:"trusted_certs"`
}

type Source interface {
	PublicSSHKeyForUsername(string) (string, error)
	Settings() (Settings, error)
}

type Blobstore struct {
	Type    string                 `json:"provider"`
	Options map[string]interface{} `json:"options"`
}

type Disks struct {
	// e.g "/dev/sda", "1"
	System string `json:"system"`

	// e.g "/dev/sdb", "2"
	Ephemeral string `json:"ephemeral"`

	// Older CPIs returned disk settings as strings
	// e.g {"disk-3845-43758-7243-38754" => "/dev/sdc"}
	//     {"disk-3845-43758-7243-38754" => "3"}
	// Newer CPIs will populate it in a hash:
	// e.g {"disk-3845-43758-7243-38754" => {"path" => "/dev/sdc"}}
	//     {"disk-3845-43758-7243-38754" => {"volume_id" => "3"}}
	Persistent map[string]interface{} `json:"persistent"`
}

type DiskSettings struct {
	ID       string
	VolumeID string
	Path     string
}

type VM struct {
	Name string `json:"name"`
}

func (s Settings) PersistentDiskSettings(diskID string) (DiskSettings, bool) {
	diskSettings := DiskSettings{}

	for id, settings := range s.Disks.Persistent {
		if id == diskID {
			diskSettings.ID = diskID

			if hashSettings, ok := settings.(map[string]interface{}); ok {
				diskSettings.Path = hashSettings["path"].(string)
				diskSettings.VolumeID = hashSettings["volume_id"].(string)
			} else {
				// Old CPIs return disk path (string) or volume id (string) as disk settings
				diskSettings.Path = settings.(string)
				diskSettings.VolumeID = settings.(string)
			}

			return diskSettings, true
		}
	}

	return diskSettings, false
}

func (s Settings) EphemeralDiskSettings() DiskSettings {
	return DiskSettings{
		VolumeID: s.Disks.Ephemeral,
		Path:     s.Disks.Ephemeral,
	}
}

type Env struct {
	Bosh BoshEnv `json:"bosh"`
}

func (e Env) GetPassword() string {
	return e.Bosh.Password
}

type BoshEnv struct {
	Password string `json:"password"`
}

type NetworkType string

const (
	NetworkTypeDynamic NetworkType = "dynamic"
	NetworkTypeVIP     NetworkType = "vip"
)

type Network struct {
	Type NetworkType `json:"type"`

	IP       string `json:"ip"`
	Netmask  string `json:"netmask"`
	Gateway  string `json:"gateway"`
	Resolved bool   `json:"resolved"` // was resolved via DHCP

	Default []string `json:"default"`
	DNS     []string `json:"dns"`

	Mac string `json:"mac"`
}

type Networks map[string]Network

func (n Networks) NetworkForMac(mac string) (Network, bool) {
	for i := range n {
		if n[i].Mac == mac {
			return n[i], true
		}
	}

	return Network{}, false
}

func (n Networks) DefaultNetworkFor(category string) (Network, bool) {
	if len(n) == 1 {
		for _, net := range n {
			return net, true
		}
	}

	for _, net := range n {
		if stringArrayContains(net.Default, category) {
			return net, true
		}
	}

	return Network{}, false
}

func stringArrayContains(stringArray []string, str string) bool {
	for _, s := range stringArray {
		if s == str {
			return true
		}
	}
	return false
}

func (n Networks) DefaultIP() (ip string, found bool) {
	for _, networkSettings := range n {
		if ip == "" {
			ip = networkSettings.IP
		}
		if len(networkSettings.Default) > 0 {
			ip = networkSettings.IP
		}
	}

	if ip != "" {
		found = true
	}
	return
}

func (n Networks) IPs() (ips []string) {
	for _, net := range n {
		if net.IP != "" {
			ips = append(ips, net.IP)
		}
	}
	return
}

func (n Network) String() string {
	return fmt.Sprintf("type: '%s', ip: '%s', netmask: '%s', gateway: '%s', mac: '%s', resolved: '%t'", n.Type, n.IP, n.Netmask, n.Gateway, n.Mac, n.Resolved)
}

func (n Network) IsDHCP() bool {
	if n.IsVIP() {
		return false
	}

	if n.isDynamic() {
		return true
	}

	// If manual network does not have IP and Netmask it cannot be statically
	// configured. We want to keep track how originally the network was resolved.
	// Otherwise it will be considered as static on subsequent checks.
	isStatic := (n.IP != "" && n.Netmask != "")
	return n.Resolved || !isStatic
}

func (n Network) isDynamic() bool {
	return n.Type == NetworkTypeDynamic
}

func (n Network) IsVIP() bool {
	return n.Type == NetworkTypeVIP
}

//{
//	"agent_id": "bm-xxxxxxxx",
//	"blobstore": {
//		"options": {
//			"blobstore_path": "/var/vcap/micro_bosh/data/cache"
//		},
//		"provider": "local"
//	},
//	"disks": {
//		"ephemeral": "/dev/sdb",
//		"persistent": {
//			"vol-xxxxxx": "/dev/sdf"
//		},
//		"system": "/dev/sda1"
//	},
//	"env": {
//		"bosh": {
//			"password": null
//		}
//	},
//  "trusted_certs": "very\nlong\nmultiline\nstring"
//	"mbus": "https://vcap:b00tstrap@0.0.0.0:6868",
//	"networks": {
//		"bosh": {
//			"cloud_properties": {
//				"subnet": "subnet-xxxxxx"
//			},
//			"default": [
//				"dns",
//				"gateway"
//			],
//			"dns": [
//				"xx.xx.xx.xx"
//			],
//			"gateway": null,
//			"ip": "xx.xx.xx.xx",
//			"netmask": null,
//			"type": "manual"
//		},
//		"vip": {
//			"cloud_properties": {},
//			"ip": "xx.xx.xx.xx",
//			"type": "vip"
//		}
//	},
//	"ntp": [
//		"0.north-america.pool.ntp.org",
//		"1.north-america.pool.ntp.org",
//		"2.north-america.pool.ntp.org",
//		"3.north-america.pool.ntp.org"
//	],
//	"vm": {
//		"name": "vm-xxxxxxxx"
//	}
//}
