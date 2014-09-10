package release

type Job struct {
	Name          string
	Fingerprint   string
	Sha1          string
	ExtractedPath string
	Templates     map[string]string
	PackageNames  []string
	Packages      []*Package
	Properties    map[string]PropertyDefinition
}

// Manifest - for reading job.MF
type Manifest struct {
	Name       string                        `yaml:"name"`
	Templates  map[string]string             `yaml:"templates"`
	Packages   []string                      `yaml:"packages"`
	Properties map[string]PropertyDefinition `yaml:"properties"`
}

type PropertyDefinition struct {
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
}

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}

	return "", false
}
