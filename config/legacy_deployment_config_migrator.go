package config

import (
	"regexp"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
)

type LegacyDeploymentConfigMigrator interface {
	Path() string
	MigrateIfExists() (migrated bool, err error)
}

type legacyDeploymentConfigMigrator struct {
	configPath              string
	deploymentConfigService DeploymentConfigService
	fs                      boshsys.FileSystem
	uuidGenerator           boshuuid.Generator
	logger                  boshlog.Logger
	logTag                  string
}

func NewLegacyDeploymentConfigMigrator(
	configPath string,
	deploymentConfigService DeploymentConfigService,
	fs boshsys.FileSystem,
	uuidGenerator boshuuid.Generator,
	logger boshlog.Logger,
) LegacyDeploymentConfigMigrator {
	return &legacyDeploymentConfigMigrator{
		configPath:              configPath,
		deploymentConfigService: deploymentConfigService,
		fs:            fs,
		uuidGenerator: uuidGenerator,
		logger:        logger,
		logTag:        "legacyDeploymentConfigMigrator",
	}
}

func (m *legacyDeploymentConfigMigrator) Path() string {
	return m.configPath
}

func (m *legacyDeploymentConfigMigrator) MigrateIfExists() (migrated bool, err error) {
	if !m.fs.FileExists(m.configPath) {
		return false, nil
	}

	deploymentConfig, err := m.migrate()
	if err != nil {
		return false, err
	}

	err = m.deploymentConfigService.Save(deploymentConfig)
	if err != nil {
		return false, bosherr.WrapError(err, "Saving migrated deployment config")
	}

	err = m.fs.RemoveAll(m.configPath)
	if err != nil {
		return false, bosherr.WrapError(err, "Deleting legacy deployment config")
	}

	return true, nil
}

func (m *legacyDeploymentConfigMigrator) migrate() (deploymentFile DeploymentFile, err error) {
	m.logger.Info(m.logTag, "Migrating legacy bosh-deployments.yml")

	bytes, err := m.fs.ReadFile(m.configPath)
	if err != nil {
		return deploymentFile, bosherr.WrapErrorf(err, "Reading legacy deployment config file '%s'", m.configPath)
	}

	// candiedyaml does not currently support ':' as the first character in a key.
	regex := regexp.MustCompile("\n([- ]) :")
	parsableString := regex.ReplaceAllString(string(bytes), "\n$1 ")

	m.logger.Debug(m.logTag, "Processed legacy bosh-deployments.yml:\n%s", parsableString)

	var legacyFile legacyDeploymentFile
	err = candiedyaml.Unmarshal([]byte(parsableString), &legacyFile)
	if err != nil {
		return deploymentFile, bosherr.WrapError(err, "Parsing job manifest")
	}

	m.logger.Debug(m.logTag, "Parsed legacy bosh-deployments.yml: %#v", legacyFile)

	uuid, err := m.uuidGenerator.Generate()
	if err != nil {
		return deploymentFile, bosherr.WrapError(err, "Generating UUID")
	}
	deploymentFile.DirectorID = uuid

	deploymentFile.Disks = []DiskRecord{}
	deploymentFile.Stemcells = []StemcellRecord{}
	deploymentFile.Releases = []ReleaseRecord{}

	if len(legacyFile.Instances) > 0 {
		instance := legacyFile.Instances[0]
		diskCID := instance.DiskCID
		if diskCID != "" {
			uuid, err = m.uuidGenerator.Generate()
			if err != nil {
				return deploymentFile, bosherr.WrapError(err, "Generating UUID")
			}

			deploymentFile.CurrentDiskID = uuid
			deploymentFile.Disks = []DiskRecord{
				{
					ID:              uuid,
					CID:             diskCID,
					Size:            0,
					CloudProperties: bmproperty.Map{},
				},
			}
		}

		vmCID := instance.VMCID
		if vmCID != "" {
			deploymentFile.CurrentVMCID = vmCID
		}

		stemcellCID := instance.StemcellCID
		if stemcellCID != "" {
			uuid, err = m.uuidGenerator.Generate()
			if err != nil {
				return deploymentFile, bosherr.WrapError(err, "Generating UUID")
			}

			stemcellName := instance.StemcellName
			if stemcellName == "" {
				stemcellName = "unknown-stemcell"
			}

			deploymentFile.Stemcells = []StemcellRecord{
				{
					ID:      uuid,
					Name:    stemcellName,
					Version: "", // unknown, will never match new stemcell
					CID:     stemcellCID,
				},
			}
		}
	}

	m.logger.Debug(m.logTag, "New deployment.json (migrated from legacy bosh-deployments.yml): %#v", deploymentFile)

	return deploymentFile, nil
}

type legacyDeploymentFile struct {
	Instances []instance `yaml:"instances"`
}

type instance struct {
	VMCID        string `yaml:"vm_cid"`
	DiskCID      string `yaml:"disk_cid"`
	StemcellCID  string `yaml:"stemcell_cid"`
	StemcellName string `yaml:"stemcell_name"`
}
