package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Manifest struct {
	Name       string
	Template   ReleaseJobRef
	Properties biproperty.Map
	Mbus       string
	Registry   Registry
}

type ReleaseJobRef struct {
	Name    string
	Release string
}

type Registry struct {
	Username  string
	Password  string
	Host      string
	Port      int
	SSHTunnel SSHTunnel
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

func (m *Manifest) PopulateRegistry(username string, password string, host string, port int, sshTunnel SSHTunnel) {
	m.Properties["registry"] = biproperty.Map{
		"host":     host,
		"port":     port,
		"username": username,
		"password": password,
	}
	m.Registry = Registry{
		Username:  username,
		Password:  password,
		Host:      host,
		Port:      port,
		SSHTunnel: sshTunnel,
	}
}

func ParseAndValidateFrom(deploymentManifestPath string, parser Parser, validator Validator, releaseSetManifest birelsetmanifest.Manifest) (Manifest, error) {
	installationSetManifest, err := parser.Parse(deploymentManifestPath)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
	}

	err = validator.Validate(installationSetManifest, releaseSetManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Validating installation manifest")
	}
	return installationSetManifest, nil
}
