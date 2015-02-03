package pkg

import (
	"sort"
)

type compilablePackages []*Package

// Sort returns a sorted shallow copy of an array of packages, in compilation order
func Sort(releasePackages []*Package) []*Package {
	sortedPackages := make(compilablePackages, len(releasePackages), len(releasePackages))
	copy(sortedPackages, releasePackages)
	sort.Sort(sortedPackages)
	return sortedPackages
}

func (p compilablePackages) Len() int {
	return len(p)
}

func (p compilablePackages) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p compilablePackages) Less(i, j int) bool {
	a := p[i]
	b := p[j]

	if len(a.Dependencies) == 0 && len(b.Dependencies) > 0 {
		return true
	}

	for _, pkg := range b.Dependencies {
		if pkg == a {
			return true
		}
	}
	return false
}
