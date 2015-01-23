package instance

import (
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type State interface {
	ToApplySpec() bmas.ApplySpec
}

type state struct {
	deploymentName      string
	name                string
	id                  int
	networks            map[string]interface{}
	jobs                []bmrel.Job
	packageBlobs        []PackageBlob
	renderedJobListBlob RenderedJobListBlob
	stateHash           string
}

// PackageBlob is a reference to the uploaded compiled package archive
type PackageBlob struct {
	Name        string
	Version     string
	SHA1        string
	BlobstoreID string
}

// RenderedJobListBlob is a reference to the uploaded rendered job templates archive
type RenderedJobListBlob struct {
	BlobstoreID string
	SHA1        string
}

func (s *state) ToApplySpec() bmas.ApplySpec {
	jobTemplates := make([]bmas.Blob, len(s.jobs), len(s.jobs))
	for i, job := range s.jobs {
		// The agent should not need the SHA1 or BlobstoreID of the release job tarball.
		// Those should only used for rendering, which we've already done.
		jobTemplates[i] = bmas.Blob{
			Name:    job.Name,
			Version: job.Fingerprint,
		}
	}

	//TODO: once packages are being compiled, create package map based on deployment + releases
	packageArchives := make(map[string]bmas.Blob, len(s.packageBlobs))
	for _, packageBlob := range s.packageBlobs {
		packageArchives[packageBlob.Name] = bmas.Blob{
			Name:        packageBlob.Name,
			Version:     packageBlob.Version,
			SHA1:        packageBlob.SHA1,
			BlobstoreID: packageBlob.BlobstoreID,
		}
	}

	return bmas.ApplySpec{
		Deployment: s.deploymentName,
		Index:      s.id,
		Networks:   s.networks,
		Job: bmas.Job{
			Name:      s.name,
			Templates: jobTemplates,
		},
		Packages: packageArchives,
		RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{
			BlobstoreID: s.renderedJobListBlob.BlobstoreID,
			SHA1:        s.renderedJobListBlob.SHA1,
		},
		ConfigurationHash: s.stateHash,
	}
}
