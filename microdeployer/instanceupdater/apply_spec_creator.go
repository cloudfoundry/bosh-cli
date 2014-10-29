package instanceupdater

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/microdeployer/agentclient"
	bmapplyspec "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type applySpecCreator struct {
	sha1Calculator bmapplyspec.SHA1Calculator
}

type ApplySpecCreator interface {
	Create(
		bmstemcell.ApplySpec,
		string,
		string,
		map[string]interface{},
		string,
		string,
		string,
	) (bmagentclient.ApplySpec, error)
}

func NewApplySpecCreator(sha1Calculator bmapplyspec.SHA1Calculator) ApplySpecCreator {
	return &applySpecCreator{
		sha1Calculator: sha1Calculator,
	}
}

func (c *applySpecCreator) Create(
	stemcellApplySpec bmstemcell.ApplySpec,
	deploymentName string,
	jobName string,
	networksSpec map[string]interface{},
	archivedTemplatesBlobID string,
	archivedTemplatesPath string,
	templatesDir string,
) (bmagentclient.ApplySpec, error) {
	archivedTemplatesSha1, err := c.sha1Calculator.Calculate(archivedTemplatesPath)
	if err != nil {
		return bmagentclient.ApplySpec{}, bosherr.WrapError(err, "Calculating archived templates SHA1")
	}

	templatesDirSha1, err := c.sha1Calculator.Calculate(templatesDir)
	if err != nil {
		return bmagentclient.ApplySpec{}, bosherr.WrapError(err, "Calculating templates dir SHA1")
	}

	applySpec := bmagentclient.ApplySpec{
		Deployment: deploymentName,
		Index:      0,
		Packages:   c.packagesSpec(stemcellApplySpec.Packages),
		Job:        c.jobSpec(stemcellApplySpec.Job.Templates, jobName),
		Networks:   networksSpec,
		RenderedTemplatesArchive: bmagentclient.RenderedTemplatesArchiveSpec{
			BlobstoreID: archivedTemplatesBlobID,
			SHA1:        archivedTemplatesSha1,
		},
		ConfigurationHash: templatesDirSha1,
	}
	return applySpec, nil
}

func (c *applySpecCreator) packagesSpec(stemcellPackages map[string]bmstemcell.Blob) map[string]bmagentclient.Blob {
	result := map[string]bmagentclient.Blob{}
	for packageName, packageBlob := range stemcellPackages {
		result[packageName] = bmagentclient.Blob{
			Name:        packageBlob.Name,
			Version:     packageBlob.Version,
			SHA1:        packageBlob.SHA1,
			BlobstoreID: packageBlob.BlobstoreID,
		}
	}

	return result
}

func (c *applySpecCreator) jobSpec(stemcellTemplates []bmstemcell.Blob, jobName string) bmagentclient.Job {
	templates := []bmagentclient.Blob{}
	for _, templateBlob := range stemcellTemplates {
		templates = append(templates, bmagentclient.Blob{
			Name:        templateBlob.Name,
			Version:     templateBlob.Version,
			SHA1:        templateBlob.SHA1,
			BlobstoreID: templateBlob.BlobstoreID,
		})
	}

	return bmagentclient.Job{
		Name:      jobName,
		Templates: templates,
	}
}
