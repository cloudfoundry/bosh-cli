package compile

import (
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type DependencyAnalysis interface {
	DeterminePackageCompilationOrder(release bmrelease.Release) ([]*bmrelease.Package, error)
}

type dependencyAnalysis struct {
	results     []*bmrelease.Package
	markedPkgs  map[*bmrelease.Package]bool
	visitedPkgs map[*bmrelease.Package]bool
}

func NewDependencyAnalylis() DependencyAnalysis {
	return &dependencyAnalysis{
		results:     []*bmrelease.Package{},
		markedPkgs:  map[*bmrelease.Package]bool{},
		visitedPkgs: map[*bmrelease.Package]bool{},
	}
}

func (da *dependencyAnalysis) DeterminePackageCompilationOrder(release bmrelease.Release) ([]*bmrelease.Package, error) {
	// Implementation of the topological sort alg outlined here http://en.wikipedia.org/wiki/Topological_sort
	for _, pkg := range release.Packages {
		da.markedPkgs[pkg] = false
	}

	pkg := da.selectUnmarked()
	for pkg != nil {
		da.visit(pkg)
		pkg = da.selectUnmarked()
	}

	return da.results, nil
}

func (da *dependencyAnalysis) visit(pkg *bmrelease.Package) {
	if da.isMarked(pkg) {
		return
	}

	if !da.isVisited(pkg) {
		da.setMark(pkg, true)
		for _, dependency := range pkg.Dependencies {
			da.visit(dependency)
		}

		da.setVisit(pkg)
		da.setMark(pkg, false)

		da.results = append(da.results, pkg)
	}
}

func (da *dependencyAnalysis) selectUnmarked() *bmrelease.Package {
	for pkg, marked := range da.markedPkgs {
		if marked == false && !da.isVisited(pkg) {
			return pkg
		}
	}
	return nil
}

func (da *dependencyAnalysis) isMarked(pkg *bmrelease.Package) bool {
	return da.markedPkgs[pkg]
}

func (da *dependencyAnalysis) isVisited(pkg *bmrelease.Package) bool {
	return da.visitedPkgs[pkg]
}

func (da *dependencyAnalysis) setMark(pkg *bmrelease.Package, marked bool) {
	da.markedPkgs[pkg] = marked
}

func (da *dependencyAnalysis) setVisit(pkg *bmrelease.Package) {
	da.visitedPkgs[pkg] = true
}
