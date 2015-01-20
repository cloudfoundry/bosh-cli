package applyspec

type ApplySpec struct {
	Deployment               string                       `json:"deployment"`
	Index                    int                          `json:"index"`
	Packages                 map[string]Blob              `json:"packages"`
	Networks                 map[string]interface{}       `json:"networks"`
	Job                      Job                          `json:"job"`
	RenderedTemplatesArchive RenderedTemplatesArchiveSpec `json:"rendered_templates_archive"`
	ConfigurationHash        string                       `json:"configuration_hash"`
}

type Blob struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	SHA1        string `json:"sha1"`
	BlobstoreID string `json:"blobstore_id"`
}

type Job struct {
	Name      string `json:"name"`
	Templates []Blob `json:"templates"`
}

type RenderedTemplatesArchiveSpec struct {
	BlobstoreID string `json:"blobstore_id"`
	SHA1        string `json:"sha1"`
}

func NewApplySpec(
	deploymentName string,
	networksSpec map[string]interface{},
	renderedTemplatesArchive TemplatesSpec,
) ApplySpec {
	return ApplySpec{
		Deployment: deploymentName,
		Index:      0,
		Networks:   networksSpec,
		RenderedTemplatesArchive: RenderedTemplatesArchiveSpec{
			BlobstoreID: renderedTemplatesArchive.BlobID,
			SHA1:        renderedTemplatesArchive.ArchiveSha1,
		},
		ConfigurationHash: renderedTemplatesArchive.ConfigurationHash,
	}
}
