package cmd

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
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

func (c SessionContextImpl) Target() string {
	if len(c.opts.TargetOpt) > 0 {
		return c.config.ResolveTarget(c.opts.TargetOpt)
	}

	return c.config.Target()
}

func (c SessionContextImpl) Credentials() cmdconf.Creds {
	creds := c.config.Credentials(c.Target())

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
	caCert := c.opts.CACertOpt

	if len(caCert) > 0 {
		if c.isFilePath(caCert) {
			readCACert, err := c.fs.ReadFileString(caCert)
			if err != nil {
				return ""
			}

			return readCACert
		}

		return caCert
	}

	return c.config.CACert(c.Target())
}

func (c SessionContextImpl) Deployment() string {
	if len(c.opts.DeploymentOpt) > 0 {
		return c.opts.DeploymentOpt
	}

	return c.config.Deployment(c.Target())
}

func (c SessionContextImpl) isFilePath(path string) bool {
	fullPath, err := c.fs.ExpandPath(path)
	if err != nil {
		return false
	}

	return c.fs.FileExists(fullPath)
}
