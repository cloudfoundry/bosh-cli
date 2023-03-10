package cmd

import (
	"fmt"
	"io/ioutil"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	semver "github.com/cppforlife/go-semi-semantic/version"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	cyclonedxjson "github.com/cloudfoundry/bosh-cli/v7/sbom/formats/cyclonedxjson"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"

	"github.com/anchore/syft/syft/artifact"
	"github.com/anchore/syft/syft/cpe"
	"github.com/anchore/syft/syft/formats"
	"github.com/anchore/syft/syft/linux"
	"github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/pkg/cataloger"
	"github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"

	"github.com/anchore/syft/syft/pkg/cataloger/dart"
	"github.com/anchore/syft/syft/pkg/cataloger/dotnet"
	"github.com/anchore/syft/syft/pkg/cataloger/elixir"
	"github.com/anchore/syft/syft/pkg/cataloger/erlang"
	"github.com/anchore/syft/syft/pkg/cataloger/golang"
	"github.com/anchore/syft/syft/pkg/cataloger/java"
	"github.com/anchore/syft/syft/pkg/cataloger/javascript"
	"github.com/anchore/syft/syft/pkg/cataloger/php"
	"github.com/anchore/syft/syft/pkg/cataloger/python"
	"github.com/anchore/syft/syft/pkg/cataloger/ruby"
	"github.com/anchore/syft/syft/pkg/cataloger/rust"
	"github.com/anchore/syft/syft/pkg/cataloger/swift"
)

type CreateSbomCmd struct {
	releaseDirFactory func(DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir)
	releaseWriter     boshrel.Writer
	compressor        boshcmd.Compressor
	fs                boshsys.FileSystem
	ui                boshui.UI
}

func NewCreateSbomCmd(
	releaseDirFactory func(DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir),
	releaseWriter boshrel.Writer,
	compressor boshcmd.Compressor,
	fs boshsys.FileSystem,
	ui boshui.UI,
) CreateSbomCmd {
	return CreateSbomCmd{releaseDirFactory, releaseWriter, compressor, fs, ui}
}

func (c CreateSbomCmd) Run(opts CreateSbomOpts) (boshrel.Release, error) {
	releaseManifestReader, releaseDir := c.releaseDirFactory(opts.Directory)
	manifestGiven := len(opts.Args.Manifest.Path) > 0

	var release boshrel.Release
	var err error
	if manifestGiven {
		release, err = releaseManifestReader.Read(opts.Args.Manifest.Path)
		if err != nil {
			return nil, err
		}
	} else {
		release, err = c.buildRelease(releaseDir, opts)
		if err != nil {
			return nil, err
		}
	}

	catalog := pkg.NewCatalog()
	relationships := []artifact.Relationship{}

	relPackageCatalog := pkg.Package{
		Name:    release.Name() + "-release",
		Version: release.Version(),
		Type:    pkg.UnknownPkg,
	}
	relPackageCatalog.OverrideID(artifact.ID(relPackageCatalog.Name))

	for _, pkg := range release.Packages() {
		extractPath, err := c.fs.TempDir("create-sbom")
		if err != nil {
			return nil, err
		}

		err = c.compressor.DecompressFileToDir(pkg.ArchivePath(), extractPath, boshcmd.CompressorOptions{})
		if err != nil {
			return nil, err
		}
		pkgCatalog, pkgRelationships, err := c.generate(extractPath, relPackageCatalog, pkg, release, opts)
		if err != nil {
			return nil, err
		}
		for p := range pkgCatalog.Enumerate() {
			catalog.Add(p)
		}
		relationships = append(relationships, pkgRelationships...)
	}

	sbom := sbom.SBOM{
		Artifacts: sbom.Artifacts{
			PackageCatalog: catalog,
		},
		Source: source.Metadata{
			Name:   release.Name() + "-release",
			Path:   release.Name() + "-release",
			Scheme: source.DirectoryScheme,
		},
		Relationships: relationships,
		Descriptor: sbom.Descriptor{
			Name:          release.Name(),
			Version:       release.Version(),
			Configuration: relPackageCatalog,
		},
	}
	encoded, _ := formats.Encode(sbom, cyclonedxjson.Format())

	//c.ui.BeginLinef("sbom to  %s\n", encoded)
	c.ui.BeginLinef("Writing sbom to %s\n", opts.Filename)
	_ = ioutil.WriteFile(opts.Filename, encoded, 0644)
	return release, nil
}

// Generate returns a populated SBOM given a path to a directory to scan.
func (c CreateSbomCmd) generate(path string, relPackageCatalog pkg.Package, boshPackage *boshpkg.Package, release boshrel.Release, opts CreateSbomOpts) (*pkg.Catalog, []artifact.Relationship, error) {
	src, err := source.NewFromDirectory(path)
	if err != nil {
		return nil, nil, err
	}

	config := cataloger.DefaultConfig()

	boshPkgCatalog := pkg.Package{
		Name:    boshPackage.Name(),
		Version: release.Version(),
		Type:    pkg.UnknownPkg,
		PURL:    fmt.Sprintf("pkg:bosh/%s/%s@%s", release.Name(), boshPackage.Name(), release.Version()),
	}

	boshPkgCatalog.SetID()

	resolver, err := src.FileResolver(config.Search.Scope)
	if err != nil {
		return nil, nil, err
	}

	relationships := []artifact.Relationship{}

	// find the distro
	distro := linux.IdentifyRelease(resolver)

	// if the catalogers have been configured, use them regardless of input type
	catalogers := []pkg.Cataloger{
		ruby.NewGemFileLockCataloger(),
		ruby.NewGemSpecCataloger(),
		python.NewPythonIndexCataloger(),
		python.NewPythonPackageCataloger(),
		javascript.NewLockCataloger(),
		javascript.NewPackageCataloger(),
		java.NewJavaCataloger(config.Java()),
		java.NewJavaPomCataloger(),
		java.NewNativeImageCataloger(),
		golang.NewGoModuleBinaryCataloger(),
		golang.NewGoModFileCataloger(),
		rust.NewCargoLockCataloger(),
		rust.NewAuditBinaryCataloger(),
		dart.NewPubspecLockCataloger(),
		dotnet.NewDotnetDepsCataloger(),
		php.NewComposerInstalledCataloger(),
		php.NewComposerLockCataloger(),
		swift.NewCocoapodsCataloger(),
		elixir.NewMixLockCataloger(),
		erlang.NewRebarLockCataloger(),
	}

	catalog, _, err := cataloger.Catalog(resolver, distro, config.Parallelism, catalogers...)
	if err != nil {
		return nil, nil, err
	}

	for p := range catalog.Enumerate() {
		catalogPackage := catalog.Package(p.ID())
		newRelationship := artifact.Relationship{
			From: catalogPackage,
			To:   &boshPkgCatalog,
			Type: artifact.DependencyOfRelationship,
		}
		relationships = append(relationships, newRelationship)
	}

	for _, packageVersion := range boshPackage.PackageVersions {
		packageCPE, err := cpe.New(fmt.Sprintf("cpe:2.3:a:%s:%s:%s:*:*:*:*:*:*:*", packageVersion.Project, packageVersion.Name, packageVersion.Version))
		if err != nil {
			return nil, nil, err
		}
		packageCatalog := pkg.Package{
			Name:    packageVersion.Name,
			Version: packageVersion.Version,
			Type:    pkg.UnknownPkg,
			PURL:    fmt.Sprintf("pkg:bosh/%s/%s@%s", release.Name(), packageVersion.Name, packageVersion.Version),
			CPEs:    []cpe.CPE{packageCPE},
		}
		packageCatalog.SetID()
		relationships = append(relationships, artifact.Relationship{
			From: &packageCatalog,
			To:   &boshPkgCatalog,
			Type: artifact.DependencyOfRelationship,
		})
		catalog.Add(packageCatalog)
	}

	newRelationship := artifact.Relationship{
		From: &boshPkgCatalog,
		To:   &relPackageCatalog,
		Type: artifact.DependencyOfRelationship,
	}
	relationships = append(relationships, newRelationship)
	catalog.Add(boshPkgCatalog)
	return catalog, relationships, nil

}

func (c CreateSbomCmd) buildRelease(releaseDir boshreldir.ReleaseDir, opts CreateSbomOpts) (boshrel.Release, error) {
	var err error

	name := opts.Name

	if len(name) == 0 {
		name, err = releaseDir.DefaultName()
		if err != nil {
			return nil, err
		}
	}

	version := semver.Version(opts.Version)

	if version.Empty() {
		version, err = releaseDir.NextDevVersion(name, opts.TimestampVersion)
		if err != nil {
			return nil, err
		}
	}

	return releaseDir.BuildRelease(name, version, opts.Force)
}
