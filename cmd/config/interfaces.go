package config

//go:generate counterfeiter . Config

type Config interface {
	Target() string
	Targets() []Target
	ResolveTarget(urlOrAlias string) string
	SetTarget(urlOrAlias, alias, caCert string) Config

	CACert(url string) string

	Credentials(url string) Creds
	SetCredentials(url string, creds Creds) Config
	UnsetCredentials(url string) Config

	Deployment(url string) string
	SetDeployment(url, nameOrPath string) Config

	Save() error
}

type Target struct {
	URL   string
	Alias string
}
