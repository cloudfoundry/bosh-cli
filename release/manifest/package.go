package manifest

type Package struct {
	Name         string   `yaml:"name"`
	Fingerprint  string   `yaml:"fingerprint"`
	Sha1         string   `yaml:"sha1"`
	Dependencies []string `yaml:"dependencies"`
}
