package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Manifest struct {
	Name       string
	Template   ReleaseJobRef
	Properties biproperty.Map
	Mbus       string
	Cert       Certificate
}

type Certificate struct {
	CA string
}

type ReleaseJobRef struct {
	Name    string
	Release string
}

type SSHTunnel struct {
	User       string
	Host       string
	Port       int
	Password   string
	PrivateKey string `yaml:"private_key"`
}
