package cmd

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
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

func (c SessionContextImpl) Environment() string {
	return c.config.ResolveEnvironment(c.opts.EnvironmentOpt)
}

func (c SessionContextImpl) Credentials() cmdconf.Creds {
	creds := c.config.Credentials(c.Environment())

	if len(c.opts.UsernameOpt) > 0 {
		creds.Username = c.opts.UsernameOpt
	}

	if len(c.opts.PasswordOpt) > 0 {
		creds.Password = c.opts.PasswordOpt
	}

	if len(c.opts.UAAClientOpt) > 0 {
		creds.Client = c.opts.UAAClientOpt
		creds.ClientSecret = c.opts.UAAClientSecretOpt
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
