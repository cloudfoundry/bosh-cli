package agentclient

type ApplySpec struct {
	Deployment               string
	Properties               map[string]interface{}
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
