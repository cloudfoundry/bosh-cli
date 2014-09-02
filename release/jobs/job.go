package jobs

type Job struct {
	Name          string
	Fingerprint   string
	Sha1          string
	ExtractedPath string
	Templates     map[string]string
	PackageNames  []string
}

func (j Job) FindTemplateByValue(value string) (string, bool) {
	for template, templateTarget := range j.Templates {
		if templateTarget == value {
			return template, true
		}
	}

	return "", false
}
