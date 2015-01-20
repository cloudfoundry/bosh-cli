package applyspec

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

// NewJob creates an applyspec job from a stemcell apply spec
func NewJob(name string, templateBlobs []bmstemcell.Blob) Job {
	templates := []Blob{}
	for _, template := range templateBlobs {
		templates = append(templates, Blob{
			Name:        template.Name,
			Version:     template.Version,
			SHA1:        template.SHA1,
			BlobstoreID: template.BlobstoreID,
		})
	}

	return Job{
		Name:      name,
		Templates: templates,
	}
}

// NewPackageMap creates an map of applyspec package blobs (indexed by name) from a stemcell apply spec
func NewPackageMap(packageBlobs map[string]bmstemcell.Blob) map[string]Blob {
	packages := map[string]Blob{}
	for packageName, packageBlob := range packageBlobs {
		packages[packageName] = Blob{
			Name:        packageBlob.Name,
			Version:     packageBlob.Version,
			SHA1:        packageBlob.SHA1,
			BlobstoreID: packageBlob.BlobstoreID,
		}
	}
	return packages
}
