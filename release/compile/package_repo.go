package compile

import (
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type packageRepo struct {
	repo map[string]*bmrelease.Package
}

func NewPackageRepo() *packageRepo {
	return &packageRepo{
		repo: make(map[string]*bmrelease.Package),
	}
}

func (pr *packageRepo) FindOrCreatePackage(pkgName string) *bmrelease.Package {
	pkg, ok := pr.repo[pkgName]
	if ok {
		return pkg
	}
	newPackage := &bmrelease.Package{
		Name: pkgName,
	}

	pr.repo[pkgName] = newPackage

	return newPackage
}
