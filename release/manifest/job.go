package manifest

type Job struct {
	Name        string `yaml:"name"`
	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`
}
