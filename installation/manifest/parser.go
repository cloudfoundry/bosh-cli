package manifest

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Parser interface {
	Parse(path string) (Manifest, error)
}

type parser struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

type manifest struct {
	Name          string
	CloudProvider installation `yaml:"cloud_provider"`
}

type installation struct {
	Properties      map[interface{}]interface{}
	Registry        Registry
	AgentEnvService string    `yaml:"agent_env_service"`
	SSHTunnel       SSHTunnel `yaml:"ssh_tunnel"`
	Mbus            string
}

func NewParser(fs boshsys.FileSystem, logger boshlog.Logger) Parser {
	return &parser{
		fs:     fs,
		logger: logger,
		logTag: "deploymentParser",
	}
}

func (p *parser) Parse(path string) (Manifest, error) {
	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	comboManifest := manifest{}
	err = candiedyaml.Unmarshal(contents, &comboManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Unmarshalling installation manifest")
	}
	p.logger.Debug(p.logTag, "Parsed installation manifest: %#v", comboManifest)

	installationManifest := p.parseInstallationManifest(comboManifest)

	return installationManifest, nil
}

func (p *parser) parseInstallationManifest(comboManifest manifest) Manifest {
	return Manifest{
		Name:            comboManifest.Name,
		Registry:        comboManifest.CloudProvider.Registry,
		AgentEnvService: comboManifest.CloudProvider.AgentEnvService,
		SSHTunnel:       comboManifest.CloudProvider.SSHTunnel,
		Mbus:            comboManifest.CloudProvider.Mbus,
		RawProperties:   comboManifest.CloudProvider.Properties,
	}
}
