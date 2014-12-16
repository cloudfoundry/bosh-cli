package applyspec

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type factory struct{}

type Factory interface {
	Create(
		bmstemcell.ApplySpec,
		string,
		string,
		map[string]interface{},
		string,
		string,
		string,
	) ApplySpec
}

func NewFactory() Factory {
	return &factory{}
}

func (c *factory) Create(
	stemcellApplySpec bmstemcell.ApplySpec,
	deploymentName string,
	jobName string,
	networksSpec map[string]interface{},
	archivedTemplatesBlobID string,
	archivedTemplatesSha1 string,
	templatesDirSha1 string,
) ApplySpec {
	applySpec := NewApplySpec(
		deploymentName,
		networksSpec,
		archivedTemplatesBlobID,
		archivedTemplatesSha1,
		templatesDirSha1,
	)
	applySpec.PopulatePackages(stemcellApplySpec.Packages)
	applySpec.PopulateJob(stemcellApplySpec.Job.Templates, jobName)

	return *applySpec
}
