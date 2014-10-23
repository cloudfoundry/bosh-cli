package stemcell

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type Stemcell struct {
	Manifest  Manifest
	ApplySpec ApplySpec
}

type Manifest struct {
	ImagePath          string
	Name               string
	Version            string
	SHA1               string
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

type ApplySpec struct {
	Job      Job
	Packages map[string]Blob
}

type Job struct {
	Name      string
	Templates []Blob
}

type Blob struct {
	Name        string
	Version     string
	SHA1        string
	BlobstoreID string `json:"blobstore_id"`
}

func (m Manifest) CloudProperties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(m.RawCloudProperties)
}
