package job

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type CompiledPackageRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}

type DependencyCompiler interface {
	Compile(releaseJobs []bmreljob.Job, stage bmui.Stage) ([]CompiledPackageRef, error)
}

type dependencyCompiler struct {
	packageCompiler bmstatepkg.Compiler
}

func NewDependencyCompiler(packageCompiler bmstatepkg.Compiler) DependencyCompiler {
	return &dependencyCompiler{
		packageCompiler: packageCompiler,
	}
}

// Compile resolves and compiles all transitive dependencies of multiple release jobs
func (c *dependencyCompiler) Compile(releaseJobs []bmreljob.Job, stage bmui.Stage) ([]CompiledPackageRef, error) {
	compileOrderReleasePackages, err := c.resolveJobCompilationDependencies(releaseJobs)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving job package dependencies")
	}

	compiledPackageRefs, err := c.compilePackages(compileOrderReleasePackages, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compiling job package dependencies")
	}

	return compiledPackageRefs, nil
}

// resolveJobPackageCompilationDependencies returns all packages required by all specified jobs, in compilation order (reverse dependency order)
func (c *dependencyCompiler) resolveJobCompilationDependencies(releaseJobs []bmreljob.Job) ([]*bmrelpkg.Package, error) {
	// collect and de-dupe all required packages (dependencies of jobs)
	nameToPackageMap := map[string]*bmrelpkg.Package{}
	for _, releaseJob := range releaseJobs {
		for _, releasePackage := range releaseJob.Packages {
			nameToPackageMap[releasePackage.Name] = releasePackage
			c.resolvePackageDependencies(releasePackage, nameToPackageMap)
		}
	}

	// flatten map values to array
	packages := make([]*bmrelpkg.Package, 0, len(nameToPackageMap))
	for _, releasePackage := range nameToPackageMap {
		packages = append(packages, releasePackage)
	}

	// sort in compilation order
	sortedPackages := bmrelpkg.Sort(packages)

	return sortedPackages, nil
}

// resolvePackageDependencies adds the releasePackage's dependencies to the nameToPackageMap recursively
func (c *dependencyCompiler) resolvePackageDependencies(releasePackage *bmrelpkg.Package, nameToPackageMap map[string]*bmrelpkg.Package) {
	for _, dependency := range releasePackage.Dependencies {
		// only add un-added packages, to avoid endless looping in case of cycles
		if _, found := nameToPackageMap[dependency.Name]; !found {
			nameToPackageMap[dependency.Name] = dependency
			c.resolvePackageDependencies(releasePackage, nameToPackageMap)
		}
	}
}

// compilePackages compiles the specified packages, in the order specified, uploads them to the Blobstore, and returns the blob references
func (c *dependencyCompiler) compilePackages(requiredPackages []*bmrelpkg.Package, stage bmui.Stage) ([]CompiledPackageRef, error) {
	packageRefs := make([]CompiledPackageRef, 0, len(requiredPackages))

	for _, pkg := range requiredPackages {
		stepName := fmt.Sprintf("Compiling package '%s/%s'", pkg.Name, pkg.Fingerprint)
		err := stage.Perform(stepName, func() error {
			compiledPackageRecord, err := c.packageCompiler.Compile(pkg)
			if err != nil {
				return err
			}

			packageRef := CompiledPackageRef{
				Name:        pkg.Name,
				Version:     pkg.Fingerprint,
				BlobstoreID: compiledPackageRecord.BlobID,
				SHA1:        compiledPackageRecord.BlobSHA1,
			}
			packageRefs = append(packageRefs, packageRef)

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return packageRefs, nil
}
