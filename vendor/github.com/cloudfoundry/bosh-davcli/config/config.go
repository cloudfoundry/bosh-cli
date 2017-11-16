package config

type Config struct {
	User          string
	Password      string
	Endpoint      string
	RetryAttempts uint
	CACert        string `json:"ca_cert"`
}
