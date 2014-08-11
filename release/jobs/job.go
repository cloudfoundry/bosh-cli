package jobs

type Job struct {
	Name          string
	Version       string
	Fingerprint   string
	Sha1          string
	ExtractedPath string
	Templates     map[string]string
	Packages      []string
}

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}

	return "", false
}
