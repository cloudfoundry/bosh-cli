package pkg

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CompiledPackage struct {
	name          string
	fingerprint   string
	osVersionSlug string

	Dependencies    []*CompiledPackage // todo privatize
	dependencyNames []string

	archivePath string
	archiveSHA1 string
}

func NewCompiledPackageWithoutArchive(name, fp, osVersionSlug, sha1 string, dependencyNames []string) *CompiledPackage {
	return &CompiledPackage{
		name:          name,
		fingerprint:   fp,
		osVersionSlug: osVersionSlug,
		archiveSHA1:   sha1,

		Dependencies:    []*CompiledPackage{},
		dependencyNames: dependencyNames,
	}
}

func NewCompiledPackageWithArchive(name, fp, osVersionSlug, path, sha1 string, dependencyNames []string) *CompiledPackage {
	return &CompiledPackage{
		name:          name,
		fingerprint:   fp,
		osVersionSlug: osVersionSlug,

		archivePath: path,
		archiveSHA1: sha1,

		Dependencies:    []*CompiledPackage{},
		dependencyNames: dependencyNames,
	}
}

func (p CompiledPackage) String() string { return p.Name() }

func (p CompiledPackage) Name() string          { return p.name }
func (p CompiledPackage) Fingerprint() string   { return p.fingerprint }
func (p CompiledPackage) OSVersionSlug() string { return p.osVersionSlug }

func (p CompiledPackage) ArchivePath() string {
	if len(p.archivePath) == 0 {
		errMsg := "Internal inconsistency: Compiled package '%s/%s' does not have archive path"
		panic(fmt.Sprintf(errMsg, p.name, p.fingerprint))
	}
	return p.archivePath
}

func (p CompiledPackage) ArchiveSHA1() string { return p.archiveSHA1 }

func (p *CompiledPackage) AttachDependencies(compiledPkgs []*CompiledPackage) error {
	for _, pkgName := range p.dependencyNames {
		var found bool

		for _, compiledPkg := range compiledPkgs {
			if compiledPkg.Name() == pkgName {
				p.Dependencies = append(p.Dependencies, compiledPkg)
				found = true
				break
			}
		}

		if !found {
			errMsg := "Expected to find compiled package '%s' since it's a dependency of compiled package '%s'"
			return bosherr.Errorf(errMsg, pkgName, p.name)
		}
	}

	return nil
}

func (p *CompiledPackage) DependencyNames() []string { return p.dependencyNames }

func (p *CompiledPackage) Deps() []Compilable {
	var coms []Compilable
	for _, dep := range p.Dependencies {
		coms = append(coms, dep)
	}
	return coms
}

func (p *CompiledPackage) IsCompiled() bool { return true }
