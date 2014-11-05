package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
)

type Cloud interface {
	bmstemcell.Infrastructure
	bmvm.Infrastructure
}

type cloud struct {
	cpiCmdRunner   CPICmdRunner
	deploymentUUID string
	logger         boshlog.Logger
	logTag         string
}

func NewCloud(
	cpiCmdRunner CPICmdRunner,
	deploymentUUID string,
	logger boshlog.Logger,
) Cloud {
	return cloud{
		cpiCmdRunner:   cpiCmdRunner,
		deploymentUUID: deploymentUUID,
		logger:         logger,
		logTag:         "cloud",
	}
}

func (c cloud) CreateStemcell(stemcellManifest bmstemcell.Manifest) (cid bmstemcell.CID, err error) {
	cloudProperties, err := stemcellManifest.CloudProperties()
	if err != nil {
		return "", bosherr.WrapError(err, "Building stemcell CloudProperties")
	}

	cmdOutput, err := c.cpiCmdRunner.Run("create_stemcell", stemcellManifest.ImagePath, cloudProperties)
	if err != nil {
		return cid, err
	}

	// for create_stemcell, the result is a string of the stemcell cid
	cidString, ok := cmdOutput.Result.(string)
	if !ok {
		return cid, bosherr.New("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}
	return bmstemcell.CID(cidString), nil
}

func (c cloud) CreateVM(
	stemcellCID bmstemcell.CID,
	cloudProperties map[string]interface{},
	networksSpec map[string]interface{},
	env map[string]interface{},
) (bmvm.CID, error) {
	diskLocality := []interface{}{} // not used with bosh-micro-cli
	cmdOutput, err := c.cpiCmdRunner.Run(
		"create_vm",
		c.deploymentUUID,
		stemcellCID,
		cloudProperties,
		networksSpec,
		diskLocality,
		env,
	)
	if err != nil {
		return "", err
	}

	// for create_vm, the result is a string of the vm cid
	cidString, ok := cmdOutput.Result.(string)
	if !ok {
		return "", bosherr.New("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}
	return bmvm.CID(cidString), nil
}
