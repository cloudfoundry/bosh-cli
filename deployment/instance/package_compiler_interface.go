package instance

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageCompiler interface {
	Compile(releasePackage *bmrel.Package, compiledPackageRefs map[string]PackageRef) (PackageRef, error)
}
