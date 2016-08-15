package template

import (
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
)

type DeploymentTemplate struct {
	template boshtpl.Template
}

func NewDeploymentTemplate(content []byte) DeploymentTemplate {
	return DeploymentTemplate{template: boshtpl.NewTemplate(content)}
}

func (t DeploymentTemplate) Evaluate(vars boshtpl.Variables) (InterpolatedTemplate, error) {
	bytes, err := t.template.Evaluate(vars)
	if err != nil {
		return InterpolatedTemplate{}, err
	}
	return NewInterpolatedTemplate(bytes), nil
}
