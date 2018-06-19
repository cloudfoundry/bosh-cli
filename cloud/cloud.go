package cloud

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

const MaxCpiApiVersionSupported = 2

// The agent on stemcells with version 2
// will avoid the registry IFF CPI and Director support CPI API v2 (above)
const StemcellNoRegistryAsOfVersion = 2
const DefaultCPIVersion = 1

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
	AttachDisk(vmCID, diskCID string) (interface{}, error)
	DetachDisk(vmCID, diskCID string) error
	DeleteDisk(diskCID string) error
	Info() (cpiInfo CpiInfo)
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
	ApiVersion      int      `json:"api_version,omitempty"`
}

type VMMetadata map[string]string

type DiskMetadata map[string]string

func NewCloud(
	cpiCmdRunner CPICmdRunner,
	directorID string,
	stemcellApiVersion int,
	logger boshlog.Logger,
) Cloud {

	cmdContext := CmdContext{
		DirectorID:         directorID,
		StemcellApiVersion: stemcellApiVersion,
	}

	logger.Debug("cloud", "Init cloud with stemcell version: %v", stemcellApiVersion)

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
	var (
		ok           bool
		cidString    string
		method       = "create_vm"
		diskLocality = []interface{}{} // not used with bosh-init
	)

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


	if c.shouldInterpretNonRegistryResponse(c.context.StemcellApiVersion, c.Info().ApiVersion) {
		var result []interface{}
		if result, ok = cmdOutput.Result.([]interface{}); ok {
			cidString, ok = result[0].(string)
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

func (c cloud) AttachDisk(vmCID, diskCID string) (interface{}, error) {
	var (
		diskHint interface{}
		method   = "attach_disk"
	)
	c.logger.Debug(c.logTag, "Attaching disk '%s' to vm '%s'", diskCID, vmCID)

	cmdOutput, err := c.cpiCmdRunner.Run(
		c.context,
		method,
		vmCID,
		diskCID,
	)

	if err != nil {
		return diskHint, bosherr.WrapError(err, "Calling CPI 'attach_disk' method")
	}

	if cmdOutput.Error != nil {
		return diskHint, NewCPIError(method, *cmdOutput.Error)
	}

	if c.shouldInterpretNonRegistryResponse(c.context.StemcellApiVersion, c.Info().ApiVersion) {
		if result, ok := cmdOutput.Result.(string); ok {
			diskHint = result
		}
	}

	return diskHint, nil
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

func (c cloud) Info() (cpiInfo CpiInfo) {
	c.logger.Debug(c.logTag, "Info")

	method := "info"
	cmdOutput, err := c.cpiCmdRunner.Run(c.context, method, " ")

	defaultResponse := CpiInfo{StemcellFormats: []string{}, ApiVersion: 1}

	if err != nil || cmdOutput.Error != nil {
		return defaultResponse
	}

	cpiInfo, err = c.infoParser(cmdOutput)
	if err != nil {
		c.logger.Debug(c.logTag, err.Error())
		return defaultResponse
	}

	c.logger.Info(c.logTag, "CPI Info resulted in %v", cpiInfo)

	return cpiInfo
}

func (c cloud) infoParser(cmdOutput CmdOutput) (cpiInfo CpiInfo, err error) {
	cpiInfo = CpiInfo{}

	data, ok := cmdOutput.Result.(map[string]interface{})
	if !ok {
		return c.raiseParsingError(fmt.Sprintf("%s", cmdOutput))
	}

	c.logger.Info(c.logTag, "CPI Info returned %v", data)

	if stemcellFormats, ok := data["stemcell_formats"].([]interface{}); ok {
		formats := []string{}
		for i := range stemcellFormats {
			formats = append(formats, stemcellFormats[i].(string))
		}
		cpiInfo.StemcellFormats = formats
	} else {
		return c.raiseParsingError(fmt.Sprintf("Extracting stemcell_formats %s", cmdOutput))
	}

	cpiInfo.ApiVersion = DefaultCPIVersion

	if apiVersion, ok := data["api_version"]; ok {
		if fltVersion, ok := apiVersion.(float64); ok {
			cpiInfo.ApiVersion = int(fltVersion)
		} else {
			return c.raiseParsingError(fmt.Sprintf("Extracting api_version %s", cmdOutput))
		}
	}

	if cpiInfo.ApiVersion > MaxCpiApiVersionSupported {
		cpiInfo.ApiVersion = MaxCpiApiVersionSupported
	}

	return cpiInfo, err
}

func (c cloud) raiseParsingError(cmdOutput string) (cpiInfo CpiInfo, err error) {
	msg := fmt.Sprintf("Unmarshalling 'info' method response failed. Result: %s", cmdOutput)
	return CpiInfo{}, bosherr.Error(msg)
}

func (c cloud) String() string {
	return fmt.Sprintf("Cloud{Context=%s}", c.context)
}

func (c cloud) shouldInterpretNonRegistryResponse(stemcellVersion int, cpiApiVersion int) bool {
	return stemcellVersion >= StemcellNoRegistryAsOfVersion && cpiApiVersion == MaxCpiApiVersionSupported
}
