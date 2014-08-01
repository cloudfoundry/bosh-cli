package config

type Service interface {
	Load() (Config, error)
	Save(Config) error
}
