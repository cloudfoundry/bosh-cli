package state

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type Builder interface {
	Build(jobName string, instanceID int, deploymentManifest bmdeplmanifest.Manifest, stage bmeventlog.Stage) (State, error)
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

func (b *builder) Build(jobName string, instanceID int, deploymentManifest bmdeplmanifest.Manifest, stage bmeventlog.Stage) (State, error) {
	deploymentJob, found := deploymentManifest.FindJobByName(jobName)
	if !found {
		return nil, bosherr.Errorf("Job '%s' not found in deployment manifest", jobName)
	}

	releaseJobs, err := b.resolveJobs(deploymentJob.Templates)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Resolving jobs for instance '%s/%d'", jobName, instanceID)
	}

	compiledPackageRefs, err := b.compileJobDependencies(releaseJobs, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Resolving job package dependencies for instance '%s/%d'", jobName, instanceID)
	}

	renderedJobTemplates, err := b.renderJobTemplates(releaseJobs, deploymentJob.Properties, deploymentManifest.Properties, deploymentManifest.Name, stage)
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

// compileJobDependencies resolves and compiles all transitive dependencies of multiple release jobs
func (b *builder) compileJobDependencies(releaseJobs []bmreljob.Job, stage bmeventlog.Stage) ([]PackageRef, error) {
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
func (b *builder) compilePackages(requiredPackages []*bmrelpkg.Package, stage bmeventlog.Stage) ([]PackageRef, error) {
	packageNamesToRefs := make(map[string]PackageRef, len(requiredPackages))

	for _, pkg := range requiredPackages {
		stepName := fmt.Sprintf("Compiling package '%s/%s'", pkg.Name, pkg.Fingerprint)
		err := stage.PerformStep(stepName, func() error {
			packageRef, err := b.packageCompiler.Compile(pkg, packageNamesToRefs)
			if err != nil {
				return err
			}
			packageNamesToRefs[packageRef.Name] = packageRef
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// flatten map values to array
	packageRefs := make([]PackageRef, 0, len(packageNamesToRefs))
	for _, packageRef := range packageNamesToRefs {
		packageRefs = append(packageRefs, packageRef)
	}

	return packageRefs, nil
}

//TODO: abstract this to a class so it can also be used by installation (to replace templates compiler)
// renderJobTemplates renders all the release job templates for multiple release jobs specified by a deployment job
func (b *builder) renderJobTemplates(releaseJobs []bmreljob.Job, jobProperties, globalProperties bmproperty.Map, deploymentName string, stage bmeventlog.Stage) (renderedJobs, error) {
	var (
		renderedJobListArchive bmtemplate.RenderedJobListArchive
		blobID                 string
	)
	err := stage.PerformStep("Rendering job templates", func() error {
		renderedJobList, err := b.jobListRenderer.Render(releaseJobs, jobProperties, globalProperties, deploymentName)
		if err != nil {
			return bosherr.WrapError(err, "Rendering job templates")
		}
		defer renderedJobList.DeleteSilently()

		renderedJobListArchive, err = b.renderedJobListCompressor.Compress(renderedJobList)
		if err != nil {
			return bosherr.WrapError(err, "Compressing rendered job templates")
		}
		defer renderedJobListArchive.DeleteSilently()

		blobID, err = b.blobstore.Add(renderedJobListArchive.Path())
		if err != nil {
			return bosherr.WrapErrorf(err, "Uploading rendered job template archive '%s' to the blobstore", renderedJobListArchive.Path())
		}

		return nil
	})
	if err != nil {
		return renderedJobs{}, err
	}

	return renderedJobs{
		BlobstoreID: blobID,
		Archive:     renderedJobListArchive,
	}, nil
}

func (b *builder) resolveJobs(jobRefs []bmdeplmanifest.ReleaseJobRef) ([]bmreljob.Job, error) {
	releaseJobs := make([]bmreljob.Job, len(jobRefs), len(jobRefs))
	for i, jobRef := range jobRefs {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.Errorf("Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
		}
		releaseJobs[i] = release
	}
	return releaseJobs, nil
}
