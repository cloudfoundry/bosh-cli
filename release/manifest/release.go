package manifest

type Manifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	CommitHash         string `yaml:"commit_hash"`
	UncommittedChanges bool   `yaml:"uncommitted_changes"`

	Jobs         []JobRef             `yaml:"jobs,omitempty"`
	Packages     []PackageRef         `yaml:"packages,omitempty"`
	CompiledPkgs []CompiledPackageRef `yaml:"compiled_packages,omitempty"`
	License      *LicenseRef          `yaml:"license,omitempty"`
}

type JobRef struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"` // todo deprecate
	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`
}

type PackageRef struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"` // todo deprecate
	Fingerprint  string   `yaml:"fingerprint"`
	SHA1         string   `yaml:"sha1"`
	Dependencies []string `yaml:"dependencies"`
}

type CompiledPackageRef struct {
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"` // todo deprecate
	Fingerprint   string   `yaml:"fingerprint"`
	SHA1          string   `yaml:"sha1"`
	OSVersionSlug string   `yaml:"stemcell"`
	Dependencies  []string `yaml:"dependencies"`
}

type LicenseRef struct {
	Version     string `yaml:"version"` // todo deprecate
	Fingerprint string `yaml:"fingerprint"`
	SHA1        string `yaml:"sha1"`
}
