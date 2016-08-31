package template

import (
	"crypto/sha512"
	"fmt"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
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

	sha_512 := sha512.New()
	_, err = sha_512.Write(bytes)
	if err != nil {
		panic("Error calculating sha_512 of interpolated template")
	}
	shaSumString := fmt.Sprintf("%x", sha_512.Sum(nil))

	return NewInterpolatedTemplate(bytes, shaSumString), nil
}
