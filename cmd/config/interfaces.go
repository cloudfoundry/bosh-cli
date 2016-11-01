package config

//go:generate counterfeiter . Config

type Config interface {
	Environments() []Environment
	ResolveEnvironment(urlOrAlias string) string
	SetEnvironment(urlOrAlias, alias, caCert string) Config

	CACert(url string) string

	Credentials(url string) Creds
	SetCredentials(url string, creds Creds) Config
	UnsetCredentials(url string) Config

	Save() error
}

type Environment struct {
	URL   string
	Alias string
}
