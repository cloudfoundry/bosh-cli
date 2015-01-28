package pkg

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageSorter interface {
	Sort([]*bmrel.Package) ([]*bmrel.Package, error)
}

type DependencyAnalysis interface {
	DeterminePackageCompilationOrder([]*bmrel.Package) ([]*bmrel.Package, error)
}

type dependencyAnalysis struct {
	results     []*bmrel.Package
	markedPkgs  map[*bmrel.Package]bool
	visitedPkgs map[*bmrel.Package]bool
}

func NewDependencyAnalysis() DependencyAnalysis {
	return &dependencyAnalysis{
		results:     []*bmrel.Package{},
		markedPkgs:  map[*bmrel.Package]bool{},
		visitedPkgs: map[*bmrel.Package]bool{},
	}
}

func (da *dependencyAnalysis) DeterminePackageCompilationOrder(packages []*bmrel.Package) ([]*bmrel.Package, error) {
	// Implementation of the topological sort alg outlined here http://en.wikipedia.org/wiki/Topological_sort
	for _, pkg := range packages {
		da.markedPkgs[pkg] = false
	}

	pkg := da.selectUnmarked()
	for pkg != nil {
		da.visit(pkg)
		pkg = da.selectUnmarked()
	}

	return da.results, nil
}

func (da *dependencyAnalysis) visit(pkg *bmrel.Package) {
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

func (da *dependencyAnalysis) selectUnmarked() *bmrel.Package {
	for pkg, marked := range da.markedPkgs {
		if marked == false && !da.isVisited(pkg) {
			return pkg
		}
	}
	return nil
}

func (da *dependencyAnalysis) isMarked(pkg *bmrel.Package) bool {
	return da.markedPkgs[pkg]
}

func (da *dependencyAnalysis) isVisited(pkg *bmrel.Package) bool {
	return da.visitedPkgs[pkg]
}

func (da *dependencyAnalysis) setMark(pkg *bmrel.Package, marked bool) {
	da.markedPkgs[pkg] = marked
}

func (da *dependencyAnalysis) setVisit(pkg *bmrel.Package) {
	da.visitedPkgs[pkg] = true
}
