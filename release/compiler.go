package release

type Compiler interface {
	Compile(Release) error
	DeterminePackageCompilationOrder(Release) ([]*Package, error)
}

type compiler struct {
}

func NewCompiler() Compiler {
	return &compiler{}
}

func (c compiler) Compile(release Release) error {
	// packageSequence, err := c.DeterminePackageCompilationOrder(release)
	// if err != nil {
	// 	return bosherr.WrapError(err, "Compiling release")
	// }

	return nil
}

// Implementation of the topological sort alg outlined here http://en.wikipedia.org/wiki/Topological_sort
func (c compiler) DeterminePackageCompilationOrder(release Release) ([]*Package, error) {
	dependencyAnalysis := newDependencyAnalylis()

	for _, pkg := range release.Packages {
		c.visit(pkg, dependencyAnalysis)
	}

	return dependencyAnalysis.results, nil
}

func (c compiler) visit(pkg *Package, da *dependencyAnalysis) {
	if da.isMarked(pkg) {
		return
	}

	if !da.isVisited(pkg) {
		da.mark(pkg, true)
		for _, dependency := range pkg.Dependencies {
			c.visit(dependency, da)
		}

		da.visit(pkg)
		da.mark(pkg, false)
		da.results = append(da.results, pkg)
	}
}

type dependencyAnalysis struct {
	results     []*Package
	markedPkgs  map[*Package]bool
	visitedPkgs map[*Package]bool
}

func newDependencyAnalylis() *dependencyAnalysis {
	return &dependencyAnalysis{
		results:     []*Package{},
		markedPkgs:  map[*Package]bool{},
		visitedPkgs: map[*Package]bool{},
	}
}

func (da *dependencyAnalysis) isMarked(pkg *Package) bool {
	return da.markedPkgs[pkg]
}

func (da *dependencyAnalysis) isVisited(pkg *Package) bool {
	return da.visitedPkgs[pkg]
}

func (da *dependencyAnalysis) mark(pkg *Package, marked bool) {
	da.markedPkgs[pkg] = marked
}

func (da *dependencyAnalysis) visit(pkg *Package) {
	da.visitedPkgs[pkg] = true
}
