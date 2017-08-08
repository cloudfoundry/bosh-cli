package config

//go:generate counterfeiter . Reader

type Reader interface {
	Read(string) (*Config, error)
}
