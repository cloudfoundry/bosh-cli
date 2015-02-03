package pkg

type packageRepo struct {
	repo map[string]*Package
}

func NewPackageRepo() *packageRepo {
	return &packageRepo{
		repo: make(map[string]*Package),
	}
}

func (pr *packageRepo) FindOrCreatePackage(pkgName string) *Package {
	pkg, ok := pr.repo[pkgName]
	if ok {
		return pkg
	}
	newPackage := &Package{
		Name: pkgName,
	}

	pr.repo[pkgName] = newPackage

	return newPackage
}
