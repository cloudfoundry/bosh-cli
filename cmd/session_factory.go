package cmd

import (
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

func NewSessionFromOpts(
	opts BoshOpts,
	config cmdconf.Config,
	ui boshui.UI,
	printEnvironment bool,
	printDeployment bool,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) Session {
	context := NewSessionContextImpl(opts, config, fs)
	clientFactory := boshhttp.NewClientFactory()

	return NewSessionImpl(context, ui, printEnvironment, printDeployment, clientFactory, logger)
}
