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
