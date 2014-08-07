package manifest

type Job struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Fingerprint string `yaml:"fingerprint"`
	Sha1        string `yaml:"sha1"`
}
