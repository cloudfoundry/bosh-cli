package template

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type DeploymentTemplateFactory interface {
	NewDeploymentTemplateFromPath(path string) (DeploymentTemplate, error)
	NewDeploymentTemplateFromBytes(bytes []byte) DeploymentTemplate
}

type templateFactory struct {
	fs boshsys.FileSystem
}

func NewDeploymentTemplateFactory(fs boshsys.FileSystem) DeploymentTemplateFactory {
	return templateFactory{fs: fs}
}

func (t templateFactory) NewDeploymentTemplateFromBytes(bytes []byte) DeploymentTemplate {
	return NewDeploymentTemplate(bytes)
}

func (t templateFactory) NewDeploymentTemplateFromPath(path string) (DeploymentTemplate, error) {
	contents, err := t.fs.ReadFile(path)
	if err != nil {
		return DeploymentTemplate{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	return t.NewDeploymentTemplateFromBytes(contents), nil
}
