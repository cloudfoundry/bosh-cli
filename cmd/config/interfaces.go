package config

import "github.com/cloudfoundry/bosh-cli/v7/uaa"

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Config

type Config interface {
	Environments() []Environment
	ResolveEnvironment(urlOrAlias string) string
	AliasEnvironment(url, alias, caCert string) (Config, error)
	UnaliasEnvironment(alias string) (Config, error)

	CACert(url string) string

	Credentials(url string) Creds
	SetCredentials(url string, creds Creds) Config
	UnsetCredentials(url string) Config
	UpdateConfigWithToken(environment string, t uaa.AccessToken) error

	Save() error
}

type Environment struct {
	URL   string
	Alias string
}
