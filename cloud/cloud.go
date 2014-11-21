package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Cloud interface {
	CreateStemcell(cloudProperties map[string]interface{}, imagePath string) (stemcellCID string, err error)
	CreateVM(
		stemcellCID string,
		cloudProperties map[string]interface{},
		networksSpec map[string]interface{},
		env map[string]interface{},
	) (vmCID string, err error)
	DeleteVM(vmCID string) error
	CreateDisk(size int, cloudProperties map[string]interface{}, vmCID string) (diskCID string, err error)
	AttachDisk(vmCID, diskCID string) error
	DetachDisk(vmCID, diskCID string) error
	DeleteDisk(diskCID string) error
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

func (c cloud) CreateDisk(size int, cloudProperties map[string]interface{}, vmCID string) (string, error) {
	c.logger.Debug(c.logTag,
		"Creating disk with size %d, cloudProperties %#v, instanceID %s",
		size,
		cloudProperties,
		vmCID,
	)
	cmdOutput, err := c.cpiCmdRunner.Run(
		"create_disk",
		size,
		cloudProperties,
		vmCID,
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

func (c cloud) DetachDisk(vmCID, diskCID string) error {
	c.logger.Debug(c.logTag, "Detaching disk '%s' from vm '%s'", diskCID, vmCID)
	_, err := c.cpiCmdRunner.Run(
		"detach_disk",
		vmCID,
		diskCID,
	)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'detach_disk' method")
	}

	return nil
}

func (c cloud) DeleteVM(vmCID string) error {
	c.logger.Debug(c.logTag, "Deleting vm '%s'", vmCID)
	_, err := c.cpiCmdRunner.Run(
		"delete_vm",
		vmCID,
	)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'delete_vm' method")
	}

	return nil
}

func (c cloud) DeleteDisk(diskCID string) error {
	c.logger.Debug(c.logTag, "Deleting disk '%s'", diskCID)
	_, err := c.cpiCmdRunner.Run(
		"delete_disk",
		diskCID,
	)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'delete_disk' method")
	}

	return nil
}
