package config

type DeploymentConfigService interface {
	Load() (DeploymentConfig, error)
	Save(DeploymentConfig) error
}
