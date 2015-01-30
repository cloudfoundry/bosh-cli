package state

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type Builder interface {
	Build(jobName string, instanceID int, deploymentManifest bmdeplmanifest.Manifest) (State, error)
}

type builder struct {
	packageCompiler           PackageCompiler
	releaseJobResolver        bmdeplrel.JobResolver
	jobListRenderer           bmtemplate.JobListRenderer
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor
	blobstore                 bmblobstore.Blobstore
	logger                    boshlog.Logger
	logTag                    string
}

func NewBuilder(
	packageCompiler PackageCompiler,
	releaseJobResolver bmdeplrel.JobResolver,
	jobListRenderer bmtemplate.JobListRenderer,
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor,
	blobstore bmblobstore.Blobstore,
	logger boshlog.Logger,
) Builder {
	return &builder{
		packageCompiler:           packageCompiler,
		releaseJobResolver:        releaseJobResolver,
		jobListRenderer:           jobListRenderer,
		renderedJobListCompressor: renderedJobListCompressor,
		blobstore:                 blobstore,
		logger:                    logger,
		logTag:                    "instanceStateBuilder",
	}
}

type renderedJobs struct {
	BlobstoreID string
	Archive     bmtemplate.RenderedJobListArchive
}

func (b *builder) Build(jobName string, instanceID int, deploymentManifest bmdeplmanifest.Manifest) (State, error) {
	deploymentJob, found := deploymentManifest.FindJobByName(jobName)
	if !found {
		return nil, bosherr.Errorf("Job '%s' not found in deployment manifest", jobName)
	}

	releaseJobs, err := b.resolveJobs(deploymentJob.Templates)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Resolving jobs for instance '%s/%d'", jobName, instanceID)
	}

	compiledPackageRefs, err := b.compileJobDependencies(releaseJobs)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Resolving job package dependencies for instance '%s/%d'", jobName, instanceID)
	}

	renderedJobTemplates, err := b.renderJobTemplates(releaseJobs, deploymentJob, deploymentManifest.Name)
	if err != nil {
		return nil, bosherr.Errorf("Rendering job templates for instance '%s/%d'", jobName, instanceID)
	}

	networkInterfaces, err := deploymentManifest.NetworkInterfaces(deploymentJob.Name)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding networks for job '%s", jobName)
	}

	// convert map to array
	networkRefs := make([]NetworkRef, 0, len(networkInterfaces))
	for networkName, networkInterface := range networkInterfaces {
		networkRefs = append(networkRefs, NetworkRef{
			Name:      networkName,
			Interface: networkInterface,
		})
	}

	// convert array to array
	renderedJobRefs := make([]JobRef, len(releaseJobs), len(releaseJobs))
	for i, releaseJob := range releaseJobs {
		renderedJobRefs[i] = JobRef{
			Name:    releaseJob.Name,
			Version: releaseJob.Fingerprint,
		}
	}

	renderedJobListArchiveBlobRef := BlobRef{
		BlobstoreID: renderedJobTemplates.BlobstoreID,
		SHA1:        renderedJobTemplates.Archive.SHA1(),
	}

	return &state{
		deploymentName:         deploymentManifest.Name,
		name:                   jobName,
		id:                     instanceID,
		networks:               networkRefs,
		renderedJobs:           renderedJobRefs,
		compiledPackages:       compiledPackageRefs,
		renderedJobListArchive: renderedJobListArchiveBlobRef,
		hash: renderedJobTemplates.Archive.Fingerprint(),
	}, nil
}

func (b *builder) compileJobDependencies(releaseJobs []bmrel.Job) ([]PackageRef, error) {
	compileOrderReleasePackages, err := b.resolveJobDependencies(releaseJobs)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving job package dependencies")
	}

	compiledPackageRefs, err := b.compilePackages(compileOrderReleasePackages)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compiling job package dependencies")
	}

	return compiledPackageRefs, nil
}

// resolveJobPackageDependencies returns all packages required by all specified jobs, in reverse dependency order (compilation order)
func (b *builder) resolveJobDependencies(releaseJobs []bmrel.Job) ([]*bmrel.Package, error) {
	// collect and de-dupe all required packages (dependencies of jobs)
	nameToPackageMap := map[string]*bmrel.Package{}
	for _, releaseJob := range releaseJobs {
		for _, releasePackage := range releaseJob.Packages {
			nameToPackageMap[releasePackage.Name] = releasePackage
			b.resolvePackageDependencies(releasePackage, nameToPackageMap)
		}
	}

	// flatten map values to array
	packages := make([]*bmrel.Package, 0, len(nameToPackageMap))
	for _, releasePackage := range nameToPackageMap {
		packages = append(packages, releasePackage)
	}

	// sort in compilation order
	sortedPackages := bmrelpkg.Sort(packages)

	return sortedPackages, nil
}

func (b *builder) resolvePackageDependencies(releasePackage *bmrel.Package, nameToPackageMap map[string]*bmrel.Package) {
	for _, dependency := range releasePackage.Dependencies {
		if _, found := nameToPackageMap[dependency.Name]; !found {
			nameToPackageMap[dependency.Name] = dependency
			b.resolvePackageDependencies(releasePackage, nameToPackageMap)
		}
	}
}

// compilePackages compiles the specified packages, in the order specified, uploads them to the Blobstore, and returns the blob references
func (b *builder) compilePackages(requiredPackages []*bmrel.Package) ([]PackageRef, error) {
	packageNamesToRefs := make(map[string]PackageRef, len(requiredPackages))
	for _, pkg := range requiredPackages {
		packageRef, err := b.packageCompiler.Compile(pkg, packageNamesToRefs)
		if err != nil {
			return []PackageRef{}, bosherr.WrapErrorf(err, "Compiling package '%s'", pkg.Name)
		}
		packageNamesToRefs[packageRef.Name] = packageRef
	}

	// flatten map values to array
	packageRefs := make([]PackageRef, 0, len(packageNamesToRefs))
	for _, packageRef := range packageNamesToRefs {
		packageRefs = append(packageRefs, packageRef)
	}

	return packageRefs, nil
}

func (b *builder) renderJobTemplates(releaseJobs []bmrel.Job, deploymentJob bmdeplmanifest.Job, deploymentName string) (renderedJobs, error) {
	jobProperties, err := deploymentJob.Properties()
	if err != nil {
		return renderedJobs{}, bosherr.WrapError(err, "Stringifying job properties")
	}

	renderedJobList, err := b.jobListRenderer.Render(releaseJobs, jobProperties, deploymentName)
	if err != nil {
		return renderedJobs{}, bosherr.WrapError(err, "Rendering job templates")
	}
	defer renderedJobList.DeleteSilently()

	renderedJobListArchive, err := b.renderedJobListCompressor.Compress(renderedJobList)
	if err != nil {
		return renderedJobs{}, bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer renderedJobListArchive.DeleteSilently()

	blobID, err := b.uploadJobTemplateListArchive(renderedJobListArchive)
	if err != nil {
		return renderedJobs{}, err
	}

	return renderedJobs{
		BlobstoreID: blobID,
		Archive:     renderedJobListArchive,
	}, nil
}

func (b *builder) uploadJobTemplateListArchive(
	renderedJobListArchive bmtemplate.RenderedJobListArchive,
) (blobID string, err error) {
	b.logger.Debug(b.logTag, "Saving job template list archive to blobstore")

	blobID, err = b.blobstore.Add(renderedJobListArchive.Path())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Uploading blob at '%s'", renderedJobListArchive.Path())
	}

	return blobID, nil
}

func (b *builder) resolveJobs(jobRefs []bmdeplmanifest.ReleaseJobRef) ([]bmrel.Job, error) {
	releaseJobs := make([]bmrel.Job, len(jobRefs), len(jobRefs))
	for i, jobRef := range jobRefs {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.Errorf("Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
		}
		releaseJobs[i] = release
	}
	return releaseJobs, nil
}
