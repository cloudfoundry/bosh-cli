package config

type UserConfigService interface {
	Load() (UserConfig, error)
	Save(UserConfig) error
}
