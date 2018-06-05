package cloud

import (
	"fmt"

	"encoding/json"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	"strings"
)

const MaxCpiApiVersionSupported = 2

type Cloud interface {
	CreateStemcell(imagePath string, cloudProperties biproperty.Map) (stemcellCID string, err error)
	DeleteStemcell(stemcellCID string) error
	HasVM(vmCID string) (bool, error)
	CreateVM(
		agentID string,
		stemcellCID string,
		cloudProperties biproperty.Map,
		networksInterfaces map[string]biproperty.Map,
		env biproperty.Map,
	) (vmCID string, err error)
	SetVMMetadata(cmCID string, metadata VMMetadata) error
	SetDiskMetadata(diskCID string, metadata DiskMetadata) error
	DeleteVM(vmCID string) error
	CreateDisk(size int, cloudProperties biproperty.Map, vmCID string) (diskCID string, err error)
	AttachDisk(vmCID, diskCID string) error
	DetachDisk(vmCID, diskCID string) error
	DeleteDisk(diskCID string) error
	Info() (cpiInfo CpiInfo, err error)
	fmt.Stringer
}

type cloud struct {
	cpiCmdRunner CPICmdRunner
	context      CmdContext
	logger       boshlog.Logger
	logTag       string
}

type CpiInfo struct {
	StemcellFormats []string `json:"stemcell_formats"`
	ApiVersion      int      `json:"api_version"`
}

type VMMetadata map[string]string

type DiskMetadata map[string]string

func NewCloud(
	cpiCmdRunner CPICmdRunner,
	directorID string,
	stemcellApiVersion int,
	logger boshlog.Logger,
) Cloud {

	cmdContext := CmdContext{DirectorID: directorID}
	if stemcellApiVersion > 0 {
		cmdContext.VM = &VM{
			Stemcell: &Stemcell{
				ApiVersion: stemcellApiVersion,
			},
		}
	}
	return cloud{
		cpiCmdRunner: cpiCmdRunner,
		context:      cmdContext,
		logger:       logger,
		logTag:       "cloud",
	}
}

func (c cloud) CreateStemcell(imagePath string, cloudProperties biproperty.Map) (string, error) {
	c.logger.Debug(c.logTag, "Creating stemcell")

	method := "create_stemcell"
	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method, imagePath, cloudProperties)
	if err != nil {
		return "", err
	}

	if cmdOutput.Error != nil {
		return "", NewCPIError(method, *cmdOutput.Error)
	}

	// for create_stemcell, the result is a string of the stemcell cid
	cidString, ok := cmdOutput.Result.(string)
	if !ok {
		return "", bosherr.Errorf("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}
	return cidString, nil
}

func (c cloud) DeleteStemcell(stemcellCID string) error {
	c.logger.Debug(c.logTag, "Deleting stemcell '%s'", stemcellCID)

	method := "delete_stemcell"
	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method, stemcellCID)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'delete_stemcell' method")
	}

	if cmdOutput.Error != nil {
		return NewCPIError(method, *cmdOutput.Error)
	}

	return nil
}

func (c cloud) HasVM(vmCID string) (bool, error) {
	method := "has_vm"
	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method, vmCID)
	if err != nil {
		return false, err
	}

	if cmdOutput.Error != nil {
		return false, NewCPIError(method, *cmdOutput.Error)
	}

	found, ok := cmdOutput.Result.(bool)
	if !ok {
		return false, bosherr.Errorf("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}
	return found, nil
}

func (c cloud) CreateVM(
	agentID string,
	stemcellCID string,
	cloudProperties biproperty.Map,
	networksInterfaces map[string]biproperty.Map,
	env biproperty.Map,
) (string, error) {
	method := "create_vm"
	diskLocality := []interface{}{} // not used with bosh-init

	cpiInfo, err := c.Info()
	if err != nil {
		return "", err
	}

	cmdOutput, err := c.cpiCmdRunner.Run(
		c.context,
		method,
		agentID,
		stemcellCID,
		cloudProperties,
		networksInterfaces,
		diskLocality,
		env,
	)
	if err != nil {
		return "", err
	}

	if cmdOutput.Error != nil {
		return "", NewCPIError(method, *cmdOutput.Error)
	}

	var ok = true
	var cidString string

	// Also consider stemcell version before interpreting the result
	if cpiInfo.ApiVersion == MaxCpiApiVersionSupported {
		result, ok := cmdOutput.Result.([]string)
		if ok {
			cidString = result[0]
		}
	} else {
		cidString, ok = cmdOutput.Result.(string)
	}

	if !ok {
		return "", bosherr.Errorf("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}

	return cidString, nil
}

func (c cloud) SetVMMetadata(vmCID string, metadata VMMetadata) error {
	cmdOutput, err := c.cpiCmdRunner.Run(
		c.context,
		"set_vm_metadata",
		vmCID,
		metadata,
	)

	if err != nil {
		return err
	}

	if cmdOutput.Error != nil {
		return NewCPIError("set_vm_metadata", *cmdOutput.Error)
	}

	return nil
}

func (c cloud) SetDiskMetadata(diskCID string, metadata DiskMetadata) error {
	cmdOutput, err := c.cpiCmdRunner.Run(
		c.context,
		"set_disk_metadata",
		diskCID,
		metadata,
	)

	if err != nil {
		return err
	}

	if cmdOutput.Error != nil {
		return NewCPIError("set_disk_metadata", *cmdOutput.Error)
	}

	return nil
}

func (c cloud) CreateDisk(size int, cloudProperties biproperty.Map, vmCID string) (string, error) {
	c.logger.Debug(c.logTag,
		"Creating disk with size %d, cloudProperties %#v, instanceID %s",
		size,
		cloudProperties,
		vmCID,
	)
	method := "create_disk"
	cmdOutput, err := c.cpiCmdRunner.Run(
		c.context,
		method,
		size,
		cloudProperties,
		vmCID,
	)
	if err != nil {
		return "", err
	}

	if cmdOutput.Error != nil {
		return "", NewCPIError(method, *cmdOutput.Error)
	}

	cidString, ok := cmdOutput.Result.(string)
	if !ok {
		return "", bosherr.Errorf("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}
	return cidString, nil
}

func (c cloud) AttachDisk(vmCID, diskCID string) error {
	c.logger.Debug(c.logTag, "Attaching disk '%s' to vm '%s'", diskCID, vmCID)
	method := "attach_disk"
	cmdOutput, err := c.cpiCmdRunner.Run(
		c.context,
		method,
		vmCID,
		diskCID,
	)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'attach_disk' method")
	}

	if cmdOutput.Error != nil {
		return NewCPIError(method, *cmdOutput.Error)
	}

	return nil
}

func (c cloud) DetachDisk(vmCID, diskCID string) error {
	c.logger.Debug(c.logTag, "Detaching disk '%s' from vm '%s'", diskCID, vmCID)
	method := "detach_disk"
	cmdOutput, err := c.cpiCmdRunner.Run(
		c.context,
		method,
		vmCID,
		diskCID,
	)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'detach_disk' method")
	}

	if cmdOutput.Error != nil {
		return NewCPIError(method, *cmdOutput.Error)
	}

	return nil
}

func (c cloud) DeleteVM(vmCID string) error {
	c.logger.Debug(c.logTag, "Deleting vm '%s'", vmCID)
	method := "delete_vm"
	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method, vmCID)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'delete_vm' method")
	}

	if cmdOutput.Error != nil {
		return NewCPIError(method, *cmdOutput.Error)
	}

	return nil
}

func (c cloud) DeleteDisk(diskCID string) error {
	c.logger.Debug(c.logTag, "Deleting disk '%s'", diskCID)
	method := "delete_disk"
	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method, diskCID)
	if err != nil {
		return bosherr.WrapError(err, "Calling CPI 'delete_disk' method")
	}

	if cmdOutput.Error != nil {
		return NewCPIError(method, *cmdOutput.Error)
	}

	return nil
}

func (c cloud) Info() (cpiInfo CpiInfo, err error) {
	c.logger.Debug(c.logTag, "Info")

	method := "info"
	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method)

	if err != nil {
		return CpiInfo{}, bosherr.WrapError(err, "Calling CPI 'info' method")
	}

	if cmdOutput.Error != nil {
		return CpiInfo{}, NewCPIError(method, *cmdOutput.Error)
	}

	cpiInfo, err = c.infoParser(cmdOutput.Result.(string))

	return cpiInfo, err
}

func (c cloud) infoParser(cmdOutput string) (cpiInfo CpiInfo, err error) {
	incoming := map[string]interface{}{}
	cpiInfo = CpiInfo{}

	err = json.Unmarshal([]byte(cmdOutput), &incoming)
	if err != nil {
		return CpiInfo{}, bosherr.WrapError(err, "Unmarshalling 'info' method response failed")
	}

	stemcellFormats, ok := incoming["stemcell_formats"].(string)
	if !ok {
		return CpiInfo{}, bosherr.Error("`stemcell_formats` must be a string")
	}
	cpiInfo.StemcellFormats = strings.Split(stemcellFormats, " ")

	version, exists := incoming["api_version"]
	if exists {
		versionFloat, ok := version.(float64)
		if !ok {
			return CpiInfo{}, bosherr.Error("`api_version` must be a number")
		}

		cpiInfo.ApiVersion = int(versionFloat)
		if cpiInfo.ApiVersion > MaxCpiApiVersionSupported {
			cpiInfo.ApiVersion = MaxCpiApiVersionSupported
		}
	}

	return cpiInfo, err
}

func (c cloud) String() string {
	return fmt.Sprintf("Cloud{Context=%s}", c.context)
}
