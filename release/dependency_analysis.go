package release

type DependencyAnalysis struct {
	results     []*Package
	markedPkgs  map[*Package]bool
	visitedPkgs map[*Package]bool
}

func NewDependencyAnalylis() *DependencyAnalysis {
	return &DependencyAnalysis{
		results:     []*Package{},
		markedPkgs:  map[*Package]bool{},
		visitedPkgs: map[*Package]bool{},
	}
}

func (da *DependencyAnalysis) DeterminePackageCompilationOrder(release Release) ([]*Package, error) {
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

func (da *DependencyAnalysis) visit(pkg *Package) {
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

func (da *DependencyAnalysis) selectUnmarked() *Package {
	for pkg, marked := range da.markedPkgs {
		if marked == false && !da.isVisited(pkg) {
			return pkg
		}
	}
	return nil
}

func (da *DependencyAnalysis) isMarked(pkg *Package) bool {
	return da.markedPkgs[pkg]
}

func (da *DependencyAnalysis) isVisited(pkg *Package) bool {
	return da.visitedPkgs[pkg]
}

func (da *DependencyAnalysis) setMark(pkg *Package, marked bool) {
	da.markedPkgs[pkg] = marked
}

func (da *DependencyAnalysis) setVisit(pkg *Package) {
	da.visitedPkgs[pkg] = true
}
