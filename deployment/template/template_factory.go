package template

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type DeploymentTemplateFactory interface {
	NewTemplateFromPath(path string) (DeploymentTemplate, error)
	NewTemplateFromBytes(bytes []byte) DeploymentTemplate
}

type templateFactory struct {
	fs boshsys.FileSystem
}

func NewTemplateFactory(fs boshsys.FileSystem) DeploymentTemplateFactory {
	return templateFactory{fs: fs}
}

func (t templateFactory) NewTemplateFromBytes(bytes []byte) DeploymentTemplate {
	return NewDeploymentTemplate(bytes)
}

func (t templateFactory) NewTemplateFromPath(path string) (DeploymentTemplate, error) {
	contents, err := t.fs.ReadFile(path)
	if err != nil {
		return DeploymentTemplate{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	return t.NewTemplateFromBytes(contents), nil
}
