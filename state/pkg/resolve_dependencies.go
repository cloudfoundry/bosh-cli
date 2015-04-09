package pkg

import (
	bmrelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
)

func ResolveDependencies(pkg *bmrelpkg.Package) []*bmrelpkg.Package {
	return resolveInner(pkg, []*bmrelpkg.Package{})
}

func resolveInner(pkg *bmrelpkg.Package, noFollow []*bmrelpkg.Package) []*bmrelpkg.Package {
	all := []*bmrelpkg.Package{}
	for _, depPkg := range pkg.Dependencies {
		if !contains(all, depPkg) && !contains(noFollow, depPkg) {
			all = append(all, depPkg)

			tDeps := resolveInner(depPkg, joinUnique(all, noFollow))
			for _, tDepPkg := range tDeps {
				all = append(all, tDepPkg)
			}
		}
	}

	for i, el := range all {
		if el == pkg {
			all = append(all[:i], all[i+1:]...)
		}
	}
	return all
}

func contains(list []*bmrelpkg.Package, element *bmrelpkg.Package) bool {
	for _, pkg := range list {
		if element == pkg {
			return true
		}
	}
	return false
}

func joinUnique(a []*bmrelpkg.Package, b []*bmrelpkg.Package) []*bmrelpkg.Package {
	joined := []*bmrelpkg.Package{}
	joined = append(joined, a...)
	for _, pkg := range b {
		if !contains(a, pkg) {
			joined = append(joined, pkg)
		}
	}
	return joined
}
