package compile

import (
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageCompiler interface {
	Compile(*bmrelease.Package) error
}
