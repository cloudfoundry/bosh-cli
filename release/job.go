package release

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type Job struct {
	Name          string
	Fingerprint   string
	SHA1          string
	ExtractedPath string
	Templates     map[string]string
	PackageNames  []string
	Packages      []*Package
	Properties    map[string]PropertyDefinition
}

type JobManifest struct {
	Name       string                        `yaml:"name"`
	Templates  map[string]string             `yaml:"templates"`
	Packages   []string                      `yaml:"packages"`
	Properties map[string]PropertyDefinition `yaml:"properties"`
}

type PropertyDefinition struct {
	Description string      `yaml:"description"`
	RawDefault  interface{} `yaml:"default"`
}

func (d PropertyDefinition) Default() (bmproperty.Property, error) {
	return bmproperty.Build(d.RawDefault)
}

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}

	return "", false
}
