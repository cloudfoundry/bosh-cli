package fakes

import (
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type FakeApplySpecFactory struct {
	CreateInput     CreateInput
	CreateApplySpec bmas.ApplySpec
}

type CreateInput struct {
	ApplySpec               bmstemcell.ApplySpec
	DeploymentName          string
	JobName                 string
	NetworksSpec            map[string]interface{}
	ArchivedTemplatesBlobID string
	ArchivedTemplatesSha1   string
	TemplatesDirSha1        string
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
	archivedTemplatesSha1 string,
	templatesDirSha1 string,
) bmas.ApplySpec {
	c.CreateInput = CreateInput{
		ApplySpec:               applySpec,
		DeploymentName:          deploymentName,
		JobName:                 jobName,
		NetworksSpec:            networksSpec,
		ArchivedTemplatesBlobID: archivedTemplatesBlobID,
		ArchivedTemplatesSha1:   archivedTemplatesSha1,
		TemplatesDirSha1:        templatesDirSha1,
	}

	return c.CreateApplySpec
}
