package config

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

/*
current_target: https://192.168.50.4:25555
targets:
- url: https://192.168.50.4:25555
  ca_cert: |...
  current_deployment: test
  username: admin
  password: admin
*/

type FSConfig struct {
	path string
	fs   boshsys.FileSystem

	schema fsConfigSchema
}

type fsConfigSchema struct {
	// Target is always a full URL
	CurrentTarget string `yaml:"current_target,omitempty"`

	Targets []fsConfigSchema_Target `yaml:"targets"`
}

type fsConfigSchema_Target struct {
	URL    string `yaml:"url"`
	CACert string `yaml:"ca_cert,omitempty"`

	Alias string `yaml:"alias,omitempty"`

	CurrentDeployment string `yaml:"current_deployment,omitempty"`

	// Auth
	Username     string `yaml:"username,omitempty"`
	Password     string `yaml:"password,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty"`
}

func NewFSConfigFromPath(path string, fs boshsys.FileSystem) (FSConfig, error) {
	var schema fsConfigSchema

	absPath, err := fs.ExpandPath(path)
	if err != nil {
		return FSConfig{}, err
	}

	if fs.FileExists(absPath) {
		bytes, err := fs.ReadFile(absPath)
		if err != nil {
			return FSConfig{}, bosherr.WrapErrorf(err, "Reading config '%s'", absPath)
		}

		err = yaml.Unmarshal(bytes, &schema)
		if err != nil {
			return FSConfig{}, bosherr.WrapError(err, "Unmarshalling config")
		}
	}

	return FSConfig{path: absPath, fs: fs, schema: schema}, nil
}

func (c FSConfig) Target() string { return c.schema.CurrentTarget }

func (c FSConfig) Targets() []Target {
	targets := []Target{}

	for _, tg := range c.schema.Targets {
		targets = append(targets, Target{URL: tg.URL, Alias: tg.Alias})
	}

	return targets
}

func (c FSConfig) ResolveTarget(urlOrAlias string) string {
	_, tg := c.findOrCreateTarget(urlOrAlias)

	return tg.URL
}

func (c FSConfig) SetTarget(urlOrAlias, alias, caCert string) Config {
	config := c.deepCopy()

	var url string

	// If url is not provided, url might actually be an alias
	if len(alias) == 0 {
		url = c.ResolveTarget(urlOrAlias)
	} else {
		url = urlOrAlias

		i, tg := config.findOrCreateTarget(url)
		tg.Alias = alias
		config.schema.Targets[i] = tg
	}

	config.schema.CurrentTarget = url

	i, tg := config.findOrCreateTarget(url)
	tg.CACert = c.readCACert(caCert)
	config.schema.Targets[i] = tg

	return config
}

func (c FSConfig) CACert(urlOrAlias string) string {
	_, tg := c.findOrCreateTarget(urlOrAlias)

	return c.readCACert(tg.CACert)
}

func (c FSConfig) readCACert(caCert string) string {
	if strings.Contains(caCert, "BEGIN") {
		return caCert
	}

	readCACert, err := c.fs.ReadFileString(caCert)
	if err != nil {
		return ""
	}

	return readCACert
}

func (c FSConfig) Credentials(urlOrAlias string) Creds {
	_, tg := c.findOrCreateTarget(urlOrAlias)

	return Creds{
		Username: tg.Username,
		Password: tg.Password,

		RefreshToken: tg.RefreshToken,
	}
}

func (c FSConfig) SetCredentials(urlOrAlias string, creds Creds) Config {
	config := c.deepCopy()

	i, tg := config.findOrCreateTarget(urlOrAlias)
	tg.Username = creds.Username
	tg.Password = creds.Password
	tg.RefreshToken = creds.RefreshToken
	config.schema.Targets[i] = tg

	return config
}

func (c FSConfig) UnsetCredentials(urlOrAlias string) Config {
	config := c.deepCopy()

	i, tg := config.findOrCreateTarget(urlOrAlias)
	tg.Username = ""
	tg.Password = ""
	tg.RefreshToken = ""
	config.schema.Targets[i] = tg

	return config
}

func (c FSConfig) Deployment(urlOrAlias string) string {
	_, tg := c.findOrCreateTarget(urlOrAlias)

	return tg.CurrentDeployment
}

func (c FSConfig) SetDeployment(urlOrAlias, nameOrPath string) Config {
	config := c.deepCopy()

	i, tg := config.findOrCreateTarget(urlOrAlias)
	tg.CurrentDeployment = nameOrPath
	config.schema.Targets[i] = tg

	return config
}

func (c FSConfig) Save() error {
	bytes, err := yaml.Marshal(c.schema)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling config")
	}

	err = c.fs.WriteFile(c.path, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing config '%s'", c.path)
	}

	return nil
}

func (c *FSConfig) findOrCreateTarget(urlOrAlias string) (int, fsConfigSchema_Target) {
	for i, tg := range c.schema.Targets {
		if urlOrAlias == tg.URL || urlOrAlias == tg.Alias {
			return i, tg
		}
	}

	tg := fsConfigSchema_Target{URL: urlOrAlias}
	c.schema.Targets = append(c.schema.Targets, tg)
	return len(c.schema.Targets) - 1, tg
}

func (c FSConfig) deepCopy() FSConfig {
	bytes, err := yaml.Marshal(c.schema)
	if err != nil {
		panic("serializing config schema")
	}

	var schema fsConfigSchema

	err = yaml.Unmarshal(bytes, &schema)
	if err != nil {
		panic("deserializing config schema")
	}

	return FSConfig{path: c.path, fs: c.fs, schema: schema}
}
