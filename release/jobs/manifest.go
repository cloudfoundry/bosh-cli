package jobs

type Manifest struct {
	Name      string            `yaml:"name"`
	Templates map[string]string `yaml:"templates"`
	Packages  []string          `yaml:"packages"`
}
