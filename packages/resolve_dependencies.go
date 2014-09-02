package packages

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

// ResolveDependencies resolves transitives package dependencies to an flattened slice
func ResolveDependencies(pkg *bmrel.Package) []*bmrel.Package {
	return resolveInner(pkg, []*bmrel.Package{})
}

// Resolves transitives package dependencies to an flattened slice
func resolveInner(pkg *bmrel.Package, noFollow []*bmrel.Package) []*bmrel.Package {
	all := []*bmrel.Package{}
	for _, depPkg := range pkg.Dependencies {
		if !contains(all, depPkg) && !contains(noFollow, depPkg) {
			all = append(all, depPkg)

			tDeps := resolveInner(depPkg, joinUnique(all, noFollow))
			for _, tDepPkg := range tDeps {
				all = append(all, tDepPkg)
			}
		}
	}
	// remove pkg if a cycle was found
	all = remove(all, pkg)
	return all
}

func remove(list []*bmrel.Package, element *bmrel.Package) []*bmrel.Package {
	for i, pkg := range list {
		if element == pkg {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func contains(list []*bmrel.Package, element *bmrel.Package) bool {
	for _, pkg := range list {
		if element == pkg {
			return true
		}
	}
	return false
}

func joinUnique(a []*bmrel.Package, b []*bmrel.Package) []*bmrel.Package {
	joined := []*bmrel.Package{}
	joined = append(joined, a...)
	for _, pkg := range b {
		if !contains(a, pkg) {
			joined = append(joined, pkg)
		}
	}
	return joined
}
