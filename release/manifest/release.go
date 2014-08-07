package manifest

type Release struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	CommitHash         string `yaml:"commit_hash"`
	UncommittedChanges bool   `yaml:"uncommitted_changes"`

	Jobs     []Job     `yaml:"jobs"`
	Packages []Package `yaml:"packages"`
}
