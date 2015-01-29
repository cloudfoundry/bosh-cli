package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type StateBuilderFactory interface {
	NewStateBuilder(bmblobstore.Blobstore) StateBuilder
}

type stateBuilderFactory struct {
	releaseJobResolver        bmdeplrel.JobResolver
	jobRenderer               bmtemplate.JobListRenderer
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor
	logger                    boshlog.Logger
}

func NewStateBuilderFactory(
	releaseJobResolver bmdeplrel.JobResolver,
	jobRenderer bmtemplate.JobListRenderer,
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor,
	logger boshlog.Logger,
) StateBuilderFactory {
	return &stateBuilderFactory{
		releaseJobResolver:        releaseJobResolver,
		jobRenderer:               jobRenderer,
		renderedJobListCompressor: renderedJobListCompressor,
		logger: logger,
	}
}

func (f *stateBuilderFactory) NewStateBuilder(blobstore bmblobstore.Blobstore) StateBuilder {
	return NewStateBuilder(
		f.releaseJobResolver,
		f.jobRenderer,
		f.renderedJobListCompressor,
		blobstore,
		f.logger,
	)
}
