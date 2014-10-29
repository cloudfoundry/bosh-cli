package applyspec

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type ApplySpec struct {
	Deployment               string
	Index                    int
	Packages                 map[string]Blob
	Networks                 map[string]interface{}
	Job                      Job
	RenderedTemplatesArchive RenderedTemplatesArchiveSpec `json:"rendered_templates_archive"`
	ConfigurationHash        string                       `json:"configuration_hash"`
}

type Blob struct {
	Name        string
	Version     string
	SHA1        string
	BlobstoreID string `json:"blobstore_id"`
}

type Job struct {
	Name      string
	Templates []Blob
}

type RenderedTemplatesArchiveSpec struct {
	BlobstoreID string `json:"blobstore_id"`
	SHA1        string
}

func NewApplySpec(
	deploymentName string,
	networksSpec map[string]interface{},
	archivedTemplatesBlobID string,
	archivedTemplatesSha1 string,
	templatesDirSha1 string,
) *ApplySpec {
	return &ApplySpec{
		Deployment: deploymentName,
		Index:      0,
		Networks:   networksSpec,
		RenderedTemplatesArchive: RenderedTemplatesArchiveSpec{
			BlobstoreID: archivedTemplatesBlobID,
			SHA1:        archivedTemplatesSha1,
		},
		ConfigurationHash: templatesDirSha1,
	}
}

func (s *ApplySpec) PopulatePackages(stemcellPackages map[string]bmstemcell.Blob) {
	packages := map[string]Blob{}
	for packageName, packageBlob := range stemcellPackages {
		packages[packageName] = Blob{
			Name:        packageBlob.Name,
			Version:     packageBlob.Version,
			SHA1:        packageBlob.SHA1,
			BlobstoreID: packageBlob.BlobstoreID,
		}
	}
	s.Packages = packages
}

func (s *ApplySpec) PopulateJob(stemcellTemplates []bmstemcell.Blob, jobName string) {
	templates := []Blob{}
	for _, templateBlob := range stemcellTemplates {
		templates = append(templates, Blob{
			Name:        templateBlob.Name,
			Version:     templateBlob.Version,
			SHA1:        templateBlob.SHA1,
			BlobstoreID: templateBlob.BlobstoreID,
		})
	}

	s.Job = Job{
		Name:      jobName,
		Templates: templates,
	}
}
