package applyspec

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type factory struct {
	sha1Calculator SHA1Calculator
}

type Factory interface {
	Create(
		bmstemcell.ApplySpec,
		string,
		string,
		map[string]interface{},
		string,
		string,
		string,
	) (ApplySpec, error)
}

func NewFactory(sha1Calculator SHA1Calculator) Factory {
	return &factory{
		sha1Calculator: sha1Calculator,
	}
}

func (c *factory) Create(
	stemcellApplySpec bmstemcell.ApplySpec,
	deploymentName string,
	jobName string,
	networksSpec map[string]interface{},
	archivedTemplatesBlobID string,
	archivedTemplatesPath string,
	templatesDir string,
) (ApplySpec, error) {
	archivedTemplatesSha1, err := c.sha1Calculator.Calculate(archivedTemplatesPath)
	if err != nil {
		return ApplySpec{}, bosherr.WrapError(err, "Calculating archived templates SHA1")
	}

	templatesDirSha1, err := c.sha1Calculator.Calculate(templatesDir)
	if err != nil {
		return ApplySpec{}, bosherr.WrapError(err, "Calculating templates dir SHA1")
	}

	applySpec := NewApplySpec(
		deploymentName,
		networksSpec,
		archivedTemplatesBlobID,
		archivedTemplatesSha1,
		templatesDirSha1,
	)
	applySpec.PopulatePackages(stemcellApplySpec.Packages)
	applySpec.PopulateJob(stemcellApplySpec.Job.Templates, jobName)

	return *applySpec, nil
}
