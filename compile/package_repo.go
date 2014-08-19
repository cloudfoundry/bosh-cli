package compile

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type packageRepo struct {
	repo map[string]*bmrel.Package
}

func NewPackageRepo() *packageRepo {
	return &packageRepo{
		repo: make(map[string]*bmrel.Package),
	}
}

func (pr *packageRepo) FindOrCreatePackage(pkgName string) *bmrel.Package {
	pkg, ok := pr.repo[pkgName]
	if ok {
		return pkg
	}
	newPackage := &bmrel.Package{
		Name: pkgName,
	}

	pr.repo[pkgName] = newPackage

	return newPackage
}
