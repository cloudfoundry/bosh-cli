package manifest

type Job struct {
	Name        string `yaml:"name"`
	Fingerprint string `yaml:"fingerprint"`
	Sha1        string `yaml:"sha1"`
}
