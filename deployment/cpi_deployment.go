package deployment

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type CPIDeployment struct {
	Name            string
	RawProperties   map[interface{}]interface{}
	Mbus            string
	Registry        Registry
	AgentEnvService string
	SSHTunnel       SSHTunnel
	Jobs            []Job
}

type Registry struct {
	Username string
	Password string
	Host     string
	Port     int
}

func (r Registry) IsEmpty() bool {
	return r == Registry{}
}

type SSHTunnel struct {
	User       string
	Host       string
	Port       int
	Password   string
	PrivateKey string `yaml:"private_key"`
}

func (d CPIDeployment) Properties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(d.RawProperties)
}
