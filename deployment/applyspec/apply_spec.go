package applyspec

// ApplySpec is the transport layer model for communicating instance state to the bosh-agent.
// The format is suboptimal for its current usage. :(
type ApplySpec struct {
	Deployment               string                       `json:"deployment"`
	Index                    int                          `json:"index"`
	Packages                 map[string]Blob              `json:"packages"`
	Networks                 map[string]interface{}       `json:"networks"`
	Job                      Job                          `json:"job"`
	RenderedTemplatesArchive RenderedTemplatesArchiveSpec `json:"rendered_templates_archive"`
	ConfigurationHash        string                       `json:"configuration_hash"`
}

// Blob is a reference to a named and versioned object, with an archive uploaded to the blobstore.
type Blob struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	SHA1        string `json:"sha1"`
	BlobstoreID string `json:"blobstore_id"`
}

// Job is a description of an instance, and the 'jobs' running on it.
// Naming uses the historical Job/Templates pattern for reverse compatibility.
// The Templates refer to release jobs to run on this instance, rendered specifically for this instance.
// The SHA/BlobstoreID of the 'Templates' are currently being ignored by the bosh-agent,
// because the RenderedTemplatesArchive contains the aggregate of all rendered jobs' templates.
// If we get to a v2 ApplySpec format, this should be flattened into the ApplySpec, with 'Templates' renamed to 'Jobs'.
type Job struct {
	Name      string `json:"name"`
	Templates []Blob `json:"templates"`
}

// RenderedTemplatesArchiveSpec is a reference to the aggregate job template archive, uploaded to the blobstore.
type RenderedTemplatesArchiveSpec struct {
	BlobstoreID string `json:"blobstore_id"`
	SHA1        string `json:"sha1"`
}
