package manifest

import (
	"github.com/cloudfoundry/bosh-agent/errors"
	version "github.com/hashicorp/go-version"
)

type ReleaseRef struct {
	Name    string
	Version string
}

func (r *ReleaseRef) VersionConstraints() (constraints version.Constraints, err error) {
	// todo: desmellify
	if r.IsLatest() {
		return nil, errors.WrapError(err, "Don't ask a 'latest' ReleaseRef for VersionConstraints")
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
