package pkg

import (
	"errors"
)

// Topologically sorts an array of packages
func Sort(releasePackages []*Package) (sortedPackages []*Package, err error){
	err = nil
	sortedPackages = []*Package{}

	incomingEdges, outgoingEdges := getEdgeMaps(releasePackages)
	noIncomingEdgesSet := []*Package{}

	for pkg, edgeList := range incomingEdges {
		if len(edgeList) == 0 {
			noIncomingEdgesSet = append(noIncomingEdgesSet, pkg)
		}
	}
	size := len(noIncomingEdgesSet)
	for size > 0 {
		elem := noIncomingEdgesSet[0]
		noIncomingEdgesSet = noIncomingEdgesSet[1:]

		sortedPackages = append([]*Package{elem}, sortedPackages...)

		for _, pkg := range outgoingEdges[elem] {
			incomingEdges[pkg] = removeFromList(incomingEdges[pkg], elem)
			if len(incomingEdges[pkg]) == 0 {
				noIncomingEdgesSet = append(noIncomingEdgesSet, pkg)
			}
		}
		size = len(noIncomingEdgesSet)
	}
	for _, edges := range incomingEdges {
		if len(edges) > 0 {
			err = errors.New("Circular dependency detected.")
		}
	}
	return
}

func removeFromList(packageList []*Package, pkg *Package) []*Package{
	for idx, elem := range packageList {
		if elem == pkg {
			packageList = append(packageList[:idx], packageList[idx+1:]...)
		}
	}
	return packageList
}

func getEdgeMaps(releasePackages []*Package) (incomingEdges, outgoingEdges  map[*Package][]*Package){
	incomingEdges = make(map[*Package][]*Package)
	outgoingEdges = make(map[*Package][]*Package)

	for _, pkg := range releasePackages {
		incomingEdges[pkg] = []*Package{}
	}

	for _, pkg := range releasePackages {
		if pkg.Dependencies != nil {
			for _, dep := range pkg.Dependencies {
				incomingEdges[dep] = append(incomingEdges[dep], pkg)
				outgoingEdges[pkg] = append(outgoingEdges[pkg], dep)
			}
		}
	}
	return
}
