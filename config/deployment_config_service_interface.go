package config

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type DeploymentFile struct {
	DirectorID          string           `json:"director_id"`
	InstallationID      string           `json:"installation_id"`
	CurrentVMCID        string           `json:"current_vm_cid"`
	CurrentStemcellID   string           `json:"current_stemcell_id"`
	CurrentDiskID       string           `json:"current_disk_id"`
	CurrentReleaseID    string           `json:"current_release_id"`
	CurrentManifestSHA1 string           `json:"current_manifest_sha1"`
	Disks               []DiskRecord     `json:"disks"`
	Stemcells           []StemcellRecord `json:"stemcells"`
	Releases            []ReleaseRecord  `json:"releases"`
}

type StemcellRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	CID     string `json:"cid"`
}

type DiskRecord struct {
	ID              string         `json:"id"`
	CID             string         `json:"cid"`
	Size            int            `json:"size"`
	CloudProperties bmproperty.Map `json:"cloud_properties"`
}

type ReleaseRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DeploymentConfigService interface {
	Exists() bool
	Load() (DeploymentFile, error)
	Save(DeploymentFile) error
}
