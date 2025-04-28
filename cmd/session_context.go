package cmd

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	cmdconf "github.com/cloudfoundry/bosh-cli/v7/cmd/config"
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
)

// SessionContextImpl prefers options over config values
type SessionContextImpl struct {
	opts   BoshOpts
	config cmdconf.Config

	fs boshsys.FileSystem
}

func NewSessionContextImpl(
	opts BoshOpts,
	config cmdconf.Config,
	fs boshsys.FileSystem,
) *SessionContextImpl {
	return &SessionContextImpl{opts: opts, config: config, fs: fs}
}

func (c SessionContextImpl) Config() cmdconf.Config {
	return c.config
}

func (c SessionContextImpl) Environment() string {
	return c.config.ResolveEnvironment(c.opts.EnvironmentOpt)
}

func (c SessionContextImpl) Credentials() cmdconf.Creds {
	creds := c.config.Credentials(c.Environment())

	if len(c.opts.ClientOpt) > 0 {
		creds.Client = c.opts.ClientOpt
		creds.ClientSecret = c.opts.ClientSecretOpt
	}

	return creds
}

func (c SessionContextImpl) CACert() string {
	if len(c.opts.CACertOpt.Content) > 0 {
		return c.opts.CACertOpt.Content
	}

	return c.config.CACert(c.Environment())
}

func (c SessionContextImpl) Deployment() string {
	return c.opts.DeploymentOpt
}
