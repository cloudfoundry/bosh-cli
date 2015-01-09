package manifest

import (
	"github.com/cloudfoundry/bosh-agent/errors"
	version "github.com/hashicorp/go-version"
)

type ReleaseRef struct {
	Name    string
	Version string
}

func NewReleaseRef(name, version string) ReleaseRef {
	return ReleaseRef{
		Name:    name,
		Version: version,
	}
}

func (r *ReleaseRef) VersionConstraints() (constraints version.Constraints, err error) {
	if r.IsLatest() {
		return nil, errors.WrapErrorf(err, "Don't ask a 'latest' ReleaseRef for VersionConstraints", r.Name)
	}

	constraints, err = version.NewConstraint(r.Version)
	if err != nil {
		return nil, errors.WrapErrorf(err, "Parsing requested version for '%s'", r.Name)
	}
	return
}

func (r *ReleaseRef) IsLatest() bool {
	return r.Version == "latest"
}
