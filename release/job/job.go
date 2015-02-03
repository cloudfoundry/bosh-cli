package job

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type Job struct {
	Name          string
	Fingerprint   string
	SHA1          string
	ExtractedPath string
	Templates     map[string]string
	PackageNames  []string
	Packages      []*bmrelpkg.Package
	Properties    map[string]PropertyDefinition
}

type PropertyDefinition struct {
	Description string
	Default     bmproperty.Property
}

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}

	return "", false
}
