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
	results := []*Package{}

	markedPkgs := map[*Package]bool{}
	visitedPkgs := map[*Package]bool{}

	for _, pkg := range release.Packages {
		c.visit(pkg, &results, &markedPkgs, &visitedPkgs)
	}

	return results, nil
}

func (c compiler) visit(pkg *Package, results *[]*Package, markedPkgs, visitedPkgs *map[*Package]bool) {
	if (*markedPkgs)[pkg] {
		return
	}

	if !(*visitedPkgs)[pkg] {
		(*markedPkgs)[pkg] = true
		for _, dependency := range pkg.Dependencies {
			c.visit(dependency, results, markedPkgs, visitedPkgs)
		}

		(*visitedPkgs)[pkg] = true
		(*markedPkgs)[pkg] = false
		*results = append(*results, pkg)
	}
}
