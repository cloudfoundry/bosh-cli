package config

type Config struct {
	User          string
	Password      string
	Endpoint      string
	RetryAttempts uint
	TLS           TLS
}

type TLS struct {
	Cert Cert
}

type Cert struct {
	CA string
}
