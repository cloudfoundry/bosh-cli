package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Cloud interface {
	CreateStemcell(map[string]interface{}, string) (string, error)
	CreateVM(string, map[string]interface{}, map[string]interface{}, map[string]interface{}) (string, error)
	CreateDisk(int, map[string]interface{}, string) (string, error)
	AttachDisk(string, string) error
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

func (c cloud) CreateStemcell(cloudProperties map[string]interface{}, imagePath string) (string, error) {
	cmdOutput, err := c.cpiCmdRunner.Run("create_stemcell", imagePath, cloudProperties)
	if err != nil {
		return "", err
	}

	// for create_stemcell, the result is a string of the stemcell cid
	cidString, ok := cmdOutput.Result.(string)
	if !ok {
		return "", bosherr.New("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}
	return cidString, nil
}

func (c cloud) CreateVM(
	stemcellCID string,
	cloudProperties map[string]interface{},
	networksSpec map[string]interface{},
	env map[string]interface{},
) (string, error) {
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
	return cidString, nil
}

func (c cloud) CreateDisk(size int, cloudProperties map[string]interface{}, instanceID string) (string, error) {
	c.logger.Debug(c.logTag,
		"Creating disk with size %d, cloudProperties %#v, instanceID %s",
		size,
		cloudProperties,
		instanceID,
	)
	cmdOutput, err := c.cpiCmdRunner.Run(
		"create_disk",
		size,
		cloudProperties,
		instanceID,
	)
	if err != nil {
		return "", err
	}

	cidString, ok := cmdOutput.Result.(string)
	if !ok {
		return "", bosherr.New("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}
	return cidString, nil
}

func (c cloud) AttachDisk(vmCID, diskCID string) error {
	c.logger.Debug(c.logTag, "Attaching disk '%s' to vm '%s'", diskCID, vmCID)
	_, err := c.cpiCmdRunner.Run(
		"attach_disk",
		vmCID,
		diskCID,
	)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'attach_disk' method")
	}

	return nil
}
