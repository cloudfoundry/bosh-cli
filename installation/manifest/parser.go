package manifest

import (
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
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
	Template        template
	Properties      map[interface{}]interface{}
	Registry        Registry
	AgentEnvService string    `yaml:"agent_env_service"`
	SSHTunnel       SSHTunnel `yaml:"ssh_tunnel"`
	Mbus            string
}

type template struct {
	Name    string
	Release string
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

	expandedPrivateKeyPath := p.expandTilde(comboManifest.CloudProvider.SSHTunnel.PrivateKey)
	comboManifest.CloudProvider.SSHTunnel.PrivateKey = expandedPrivateKeyPath

	installationManifest := Manifest{
		Name: comboManifest.Name,
		Template: ReleaseJobRef{
			Name:    comboManifest.CloudProvider.Template.Name,
			Release: comboManifest.CloudProvider.Template.Release,
		},
		Registry:        comboManifest.CloudProvider.Registry,
		AgentEnvService: comboManifest.CloudProvider.AgentEnvService,
		SSHTunnel:       comboManifest.CloudProvider.SSHTunnel,
		Mbus:            comboManifest.CloudProvider.Mbus,
	}

	properties, err := bmproperty.BuildMap(comboManifest.CloudProvider.Properties)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing cloud_provider manifest properties: %#v", comboManifest.CloudProvider.Properties)
	}
	installationManifest.Properties = properties

	return installationManifest, nil
}

// special case handling for linux/darwin where the tilde character resolves to home
func (p *parser) expandTilde(rawPath string) string {
	currentUser, err := user.Current()
	if err != nil {
		p.logger.Warn(p.logTag, "Unable to get current user, cannot expand '~' in paths")
		return rawPath
	}

	if strings.IndexRune(rawPath, '~') != 0 {
		return rawPath
	}

	sep := "~" + string(os.PathSeparator)
	currentUserHome := currentUser.HomeDir + string(os.PathSeparator) // we'll clean up any extra path separators later
	expandedPath := path.Clean(strings.Replace(rawPath, sep, currentUserHome, 1))
	return path.Clean(expandedPath)
}
