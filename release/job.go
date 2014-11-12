package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
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

func (d PropertyDefinition) Default() (interface{}, error) {
	defaultMap, ok := d.RawDefault.(map[interface{}]interface{})
	if ok {
		stringifiedMap, err := bmkeystr.NewKeyStringifier().ConvertMap(defaultMap)
		if err != nil {
			return nil, bosherr.WrapError(err, "Converting job manifest properties")
		}
		return stringifiedMap, nil
	}
	return d.RawDefault, nil
}

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}

	return "", false
}
