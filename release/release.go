package release

import (
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type Release struct {
	Name    string
	Version string

	CommitHash         string
	UncommittedChanges bool

	Jobs          []bmreljob.Job
	Packages      []Package
	ExtractedPath string
}

type Package struct {
	Name          string
	Version       string
	Fingerprint   string
	Sha1          string
	Dependencies  []string
	ExtractedPath string
}
