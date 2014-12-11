package deployment

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
)

type CPIDeploymentManifest struct {
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

func (o SSHTunnel) IsEmpty() bool {
	return o == SSHTunnel{}
}

func (d CPIDeploymentManifest) Properties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(d.RawProperties)
}

type CPIDeployment interface {
	StartRegistry() (bmregistry.Server, error)
}

type cpiDeployment struct {
	manifest              CPIDeploymentManifest
	registryServerManager bmregistry.ServerManager
}

func NewCPIDeployment(
	manifest CPIDeploymentManifest,
	registryServerManager bmregistry.ServerManager,
) CPIDeployment {
	return &cpiDeployment{
		manifest:              manifest,
		registryServerManager: registryServerManager,
	}
}

func (d *cpiDeployment) Manifest() CPIDeploymentManifest {
	return d.manifest
}

func (d *cpiDeployment) StartRegistry() (bmregistry.Server, error) {
	config := d.manifest.Registry
	return d.registryServerManager.Start(config.Username, config.Password, config.Host, config.Port)
}
