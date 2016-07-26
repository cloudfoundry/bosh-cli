package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshuaa "github.com/cloudfoundry/bosh-init/uaa"
)

//go:generate counterfeiter . SessionContext

type SessionContext interface {
	Environment() string
	CACert() string
	Credentials() cmdconf.Creds

	Deployment() string
}

//go:generate counterfeiter . Session

type Session interface {
	Environment() string
	Credentials() cmdconf.Creds

	UAA() (boshuaa.UAA, error)

	Director() (boshdir.Director, error)
	AnonymousDirector() (boshdir.Director, error)

	Deployment() (boshdir.Deployment, error)
}
