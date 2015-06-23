package job

import (
	biproperty "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/property"
	birelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
)

type Job struct {
	Name          string
	Fingerprint   string
	SHA1          string
	ExtractedPath string
	Templates     map[string]string
	PackageNames  []string
	Packages      []*birelpkg.Package
	Properties    map[string]PropertyDefinition
}

type PropertyDefinition struct {
	Description string
	Default     biproperty.Property
}

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}

	return "", false
}
