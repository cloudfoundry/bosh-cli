package config

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type DeploymentState struct {
	DirectorID         string           `json:"director_id"`
	InstallationID     string           `json:"installation_id"`
	CurrentVMCID       string           `json:"current_vm_cid"`
	CurrentStemcellID  string           `json:"current_stemcell_id"`
	CurrentDiskID      string           `json:"current_disk_id"`
	CurrentReleaseIDs  []string         `json:"current_release_ids"`
	CurrentManifestSHA string           `json:"current_manifest_sha"`
	Disks              []DiskRecord     `json:"disks"`
	Stemcells          []StemcellRecord `json:"stemcells"`
	Releases           []ReleaseRecord  `json:"releases"`
	CurrentVMs         []VMRecord       `json:"current_vms,omitempty"`
}

// VMRecord stores per-instance VM state: cloud identity, disk association,
// the static IP used to reconstruct the per-instance agent mbus URL at runtime,
// and the availability zone the VM was placed in (for sticky re-deployment).
type VMRecord struct {
	ID            string `json:"id"`
	JobName       string `json:"job_name"`
	InstanceID    int    `json:"instance_id"`
	CID           string `json:"cid,omitempty"`
	CurrentDiskID string `json:"current_disk_id,omitempty"`
	StaticIP      string `json:"static_ip,omitempty"`
	AZ            string `json:"az,omitempty"`
}

type StemcellRecord struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	ApiVersion int    `json:"api_version,omitempty"`
	CID        string `json:"cid"`
}

type DiskRecord struct {
	ID              string         `json:"id"`
	CID             string         `json:"cid"`
	Size            int            `json:"size"`
	CloudProperties biproperty.Map `json:"cloud_properties"`
}

type ReleaseRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DeploymentStateService interface {
	Path() string
	Exists() bool
	Load() (DeploymentState, error)
	Save(DeploymentState) error
	Cleanup() error
}
