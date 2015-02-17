package state

import (
	"fmt"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Builder interface {
	Build(bminstallmanifest.Manifest, bmui.Stage) (State, error)
}

type builder struct {
	releaseJobResolver bmdeplrel.JobResolver
	packageCompiler    bmstatepkg.Compiler
	jobListRenderer    bmtemplate.JobListRenderer
	compressor         boshcmd.Compressor
	blobstore          boshblob.Blobstore
	templatesRepo      bmtemplate.TemplatesRepo
}

func NewBuilder(
	releaseJobResolver bmdeplrel.JobResolver,
	packageCompiler bmstatepkg.Compiler,
	jobListRenderer bmtemplate.JobListRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo bmtemplate.TemplatesRepo,
) Builder {
	return &builder{
		releaseJobResolver: releaseJobResolver,
		packageCompiler:    packageCompiler,
		jobListRenderer:    jobListRenderer,
		compressor:         compressor,
		blobstore:          blobstore,
		templatesRepo:      templatesRepo,
	}
}

func (b *builder) Build(installationManifest bminstallmanifest.Manifest, stage bmui.Stage) (State, error) {
	// installation only ever has one job: the cpi job.
	releaseJobRefs := []bminstallmanifest.ReleaseJobRef{installationManifest.Template}

	// installation jobs do not get rendered with global deployment properties, only the cloud_provider properties
	globalProperties := bmproperty.Map{}
	jobProperties := installationManifest.Properties

	//TODO: bellow is similar to deployment/instance/state.Builder - abstract?
	releaseJobs, err := b.resolveJobs(releaseJobRefs)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving jobs for installation")
	}

	compiledPackageRefs, err := b.compileJobDependencies(releaseJobs, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving job package dependencies for installation")
	}

	renderedJobRefs, err := b.renderJobTemplates(releaseJobs, jobProperties, globalProperties, installationManifest.Name, stage)
	if err != nil {
		return nil, bosherr.Error("Rendering job templates for installation")
	}

	if len(renderedJobRefs) != 1 {
		return nil, bosherr.Error("Too many jobs rendered... oops?")
	}

	return NewState(renderedJobRefs[0], compiledPackageRefs), nil
}

//TODO: similar to deployment/instance/state.Builder - abstract?
func (b *builder) resolveJobs(jobRefs []bminstallmanifest.ReleaseJobRef) ([]bmreljob.Job, error) {
	releaseJobs := make([]bmreljob.Job, len(jobRefs), len(jobRefs))
	for i, jobRef := range jobRefs {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.WrapErrorf(err, "Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
		}
		releaseJobs[i] = release
	}
	return releaseJobs, nil
}

//TODO: same as deployment/instance/state.Builder - abstract
// compileJobDependencies resolves and compiles all transitive dependencies of multiple release jobs
func (b *builder) compileJobDependencies(releaseJobs []bmreljob.Job, stage bmui.Stage) ([]bminstallpkg.CompiledPackageRef, error) {
	compileOrderReleasePackages, err := b.resolveJobCompilationDependencies(releaseJobs)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving job package dependencies")
	}

	compiledPackageRefs, err := b.compilePackages(compileOrderReleasePackages, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compiling job package dependencies")
	}

	return compiledPackageRefs, nil
}

//TODO: same as deployment/instance/state.Builder - abstract
// resolveJobPackageCompilationDependencies returns all packages required by all specified jobs, in compilation order (reverse dependency order)
func (b *builder) resolveJobCompilationDependencies(releaseJobs []bmreljob.Job) ([]*bmrelpkg.Package, error) {
	// collect and de-dupe all required packages (dependencies of jobs)
	nameToPackageMap := map[string]*bmrelpkg.Package{}
	for _, releaseJob := range releaseJobs {
		for _, releasePackage := range releaseJob.Packages {
			nameToPackageMap[releasePackage.Name] = releasePackage
			b.resolvePackageDependencies(releasePackage, nameToPackageMap)
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

//TODO: same as deployment/instance/state.Builder - abstract
// resolvePackageDependencies adds the releasePackage's dependencies to the nameToPackageMap recursively
func (b *builder) resolvePackageDependencies(releasePackage *bmrelpkg.Package, nameToPackageMap map[string]*bmrelpkg.Package) {
	for _, dependency := range releasePackage.Dependencies {
		// only add un-added packages, to avoid endless looping in case of cycles
		if _, found := nameToPackageMap[dependency.Name]; !found {
			nameToPackageMap[dependency.Name] = dependency
			b.resolvePackageDependencies(releasePackage, nameToPackageMap)
		}
	}
}

// compilePackages compiles the specified packages, in the order specified, uploads them to the Blobstore, and returns the blob references
func (b *builder) compilePackages(requiredPackages []*bmrelpkg.Package, stage bmui.Stage) ([]bminstallpkg.CompiledPackageRef, error) {
	packageNamesToRefs := make(map[string]bminstallpkg.CompiledPackageRef, len(requiredPackages))

	for _, pkg := range requiredPackages {
		stepName := fmt.Sprintf("Compiling package '%s/%s'", pkg.Name, pkg.Fingerprint)
		err := stage.Perform(stepName, func() error {

			compiledPackageRecord, err := b.packageCompiler.Compile(pkg)
			if err != nil {
				return err
			}

			packageNamesToRefs[pkg.Name] = bminstallpkg.CompiledPackageRef{
				Name:        pkg.Name,
				Version:     pkg.Fingerprint,
				BlobstoreID: compiledPackageRecord.BlobID,
				SHA1:        compiledPackageRecord.BlobSHA1,
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// flatten map values to array
	packageRefs := make([]bminstallpkg.CompiledPackageRef, 0, len(packageNamesToRefs))
	for _, packageRef := range packageNamesToRefs {
		packageRefs = append(packageRefs, packageRef)
	}

	return packageRefs, nil
}

// renderJobTemplates renders all the release job templates for multiple release jobs specified by a deployment job
func (b *builder) renderJobTemplates(
	releaseJobs []bmreljob.Job,
	jobProperties bmproperty.Map,
	globalProperties bmproperty.Map,
	deploymentName string,
	stage bmui.Stage,
) ([]bminstalljob.RenderedJobRef, error) {
	renderedJobRefs := make([]bminstalljob.RenderedJobRef, 0, len(releaseJobs))
	err := stage.Perform("Rendering job templates", func() error {
		renderedJobList, err := b.jobListRenderer.Render(releaseJobs, jobProperties, globalProperties, deploymentName)
		if err != nil {
			return err
		}
		defer renderedJobList.DeleteSilently()

		for _, renderedJob := range renderedJobList.All() {
			renderedJobRecord, err := b.compressAndUpload(renderedJob)
			if err != nil {
				return err
			}

			releaseJob := renderedJob.Job()
			renderedJobRefs = append(renderedJobRefs, bminstalljob.RenderedJobRef{
				Name:        releaseJob.Name,
				Version:     releaseJob.Fingerprint,
				BlobstoreID: renderedJobRecord.BlobID,
				SHA1:        renderedJobRecord.BlobSHA1,
			})
		}

		return nil
	})

	return renderedJobRefs, err
}

func (b *builder) compressAndUpload(renderedJob bmtemplate.RenderedJob) (record bmtemplate.TemplateRecord, err error) {
	tarballPath, err := b.compressor.CompressFilesInDir(renderedJob.Path())
	if err != nil {
		return record, bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer b.compressor.CleanUp(tarballPath)

	blobID, blobSHA1, err := b.blobstore.Create(tarballPath)
	if err != nil {
		return record, bosherr.WrapError(err, "Creating blob")
	}

	record = bmtemplate.TemplateRecord{
		BlobID:   blobID,
		BlobSHA1: blobSHA1,
	}
	//TODO: move TemplatesRepo to state/job.TemplatesRepo and reuse in deployment/instance/state.Builder
	err = b.templatesRepo.Save(renderedJob.Job(), record)
	if err != nil {
		return record, bosherr.WrapError(err, "Saving job to templates repo")
	}

	return record, nil
}
