package fakes

import (
	bmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type FakeApplySpecFactory struct {
	CreateInput     CreateInput
	CreateApplySpec bmas.ApplySpec
	CreateErr       error
}

type CreateInput struct {
	ApplySpec               bmstemcell.ApplySpec
	DeploymentName          string
	JobName                 string
	NetworksSpec            map[string]interface{}
	ArchivedTemplatesBlobID string
	ArchivedTemplatesPath   string
	TemplatesDir            string
}

func NewFakeApplySpecFactory() *FakeApplySpecFactory {
	return &FakeApplySpecFactory{}
}

func (c *FakeApplySpecFactory) Create(
	applySpec bmstemcell.ApplySpec,
	deploymentName string,
	jobName string,
	networksSpec map[string]interface{},
	archivedTemplatesBlobID string,
	archivedTemplatesPath string,
	templatesDir string,
) (bmas.ApplySpec, error) {
	c.CreateInput = CreateInput{
		ApplySpec:               applySpec,
		DeploymentName:          deploymentName,
		JobName:                 jobName,
		NetworksSpec:            networksSpec,
		ArchivedTemplatesBlobID: archivedTemplatesBlobID,
		ArchivedTemplatesPath:   archivedTemplatesPath,
		TemplatesDir:            templatesDir,
	}

	return c.CreateApplySpec, c.CreateErr
}
