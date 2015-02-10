package state

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type BuilderFactory interface {
	NewBuilder(bmblobstore.Blobstore, bmagentclient.AgentClient) Builder
}

type builderFactory struct {
	packageRepo               bmstatepkg.CompiledPackageRepo
	releaseJobResolver        bmdeplrel.JobResolver
	jobRenderer               bmtemplate.JobListRenderer
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor
	logger                    boshlog.Logger
}

func NewBuilderFactory(
	packageRepo bmstatepkg.CompiledPackageRepo,
	releaseJobResolver bmdeplrel.JobResolver,
	jobRenderer bmtemplate.JobListRenderer,
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor,
	logger boshlog.Logger,
) BuilderFactory {
	return &builderFactory{
		packageRepo:               packageRepo,
		releaseJobResolver:        releaseJobResolver,
		jobRenderer:               jobRenderer,
		renderedJobListCompressor: renderedJobListCompressor,
		logger: logger,
	}
}

func (f *builderFactory) NewBuilder(blobstore bmblobstore.Blobstore, agentClient bmagentclient.AgentClient) Builder {
	packageCompiler := NewRemotePackageCompiler(blobstore, agentClient, f.packageRepo)
	return NewBuilder(
		packageCompiler,
		f.releaseJobResolver,
		f.jobRenderer,
		f.renderedJobListCompressor,
		blobstore,
		f.logger,
	)
}
