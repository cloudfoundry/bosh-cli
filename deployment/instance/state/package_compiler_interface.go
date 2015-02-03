package state

import (
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type PackageCompiler interface {
	Compile(releasePackage *bmrelpkg.Package, compiledPackageRefs map[string]PackageRef) (PackageRef, error)
}
