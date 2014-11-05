package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type CmdContext struct {
	DirectorUUID string `json:"director_uuid"`
}

type CmdInput struct {
	Method    string        `json:"method"`
	Arguments []interface{} `json:"arguments"`
	Context   CmdContext    `json:"context"`
}

type CmdError struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	OkToRetry bool   `json:"ok_to_retry"`
}

func (e CmdError) String() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		panic(fmt.Sprintf("Error stringifying CmdError %#v: %s", e, err.Error()))
	}
	return fmt.Sprintf("CmdError%s", string(bytes))
}

type CmdOutput struct {
	Result interface{} `json:"result"`
	Error  *CmdError   `json:"error,omitempty"`
	Log    string      `json:"log"`
}

type CPIJob struct {
	JobPath      string
	JobsPath     string
	PackagesPath string
}

func (j CPIJob) String() string {
	return fmt.Sprintf(
		"CPIJob{JobPath:'%s', JobsPath:'%s', PackagesPath:'%s'}",
		j.JobPath,
		j.JobsPath,
		j.PackagesPath,
	)
}

type Cloud interface {
	bmstemcell.Infrastructure
	bmvm.Infrastructure
}

type cloud struct {
	fs             boshsys.FileSystem
	cmdRunner      boshsys.CmdRunner
	cpiJob         CPIJob
	deploymentUUID string
	logger         boshlog.Logger
	logTag         string
}

func NewCloud(
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	cpiJob CPIJob,
	deploymentUUID string,
	logger boshlog.Logger,
) Cloud {
	return cloud{
		fs:             fs,
		cmdRunner:      cmdRunner,
		cpiJob:         cpiJob,
		deploymentUUID: deploymentUUID,
		logger:         logger,
		logTag:         "cloud",
	}
}

func (c cloud) String() string {
	return fmt.Sprintf(
		"Cloud{cpiJob:%s, deploymentUUID:'%s', logTag:'%s'}",
		c.cpiJob,
		c.deploymentUUID,
		c.logTag,
	)
}

func (c cloud) CreateStemcell(stemcellManifest bmstemcell.Manifest) (cid bmstemcell.CID, err error) {
	cloudProperties, err := stemcellManifest.CloudProperties()
	if err != nil {
		return "", bosherr.WrapError(err, "Building stemcell CloudProperties")
	}

	cmdOutput, err := c.execCPICmd("create_stemcell", stemcellManifest.ImagePath, cloudProperties)
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
	cmdOutput, err := c.execCPICmd(
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

func (c cloud) cpiExecutablePath() string {
	return filepath.Join(c.cpiJob.JobPath, "bin", "cpi")
}

func (c cloud) execCPICmd(method string, args ...interface{}) (CmdOutput, error) {
	cmdInput := CmdInput{
		Method:    method,
		Arguments: args,
		Context: CmdContext{
			DirectorUUID: c.deploymentUUID,
		},
	}
	inputBytes, err := json.Marshal(cmdInput)
	if err != nil {
		return CmdOutput{}, bosherr.WrapError(err, "Marshalling external CPI command input %#v", cmdInput)
	}

	cmdPath := c.cpiExecutablePath()
	cmd := boshsys.Command{
		Name: cmdPath,
		Env: map[string]string{
			"BOSH_PACKAGES_DIR": c.cpiJob.PackagesPath,
			"BOSH_JOBS_DIR":     c.cpiJob.JobsPath,
			"PATH":              "/usr/local/bin:/usr/bin:/bin",
		},
		UseIsolatedEnv: true,
		Stdin:          bytes.NewReader(inputBytes),
	}
	stdout, stderr, exitCode, err := c.cmdRunner.RunComplexCommand(cmd)
	c.logger.Debug(c.logTag, "Exit Code %d when executing external CPI command '%s'\nSTDIN: '%s'\nSTDOUT: '%s'\nSTDERR: '%s'", exitCode, cmdPath, string(inputBytes), stdout, stderr)
	if err != nil {
		return CmdOutput{}, bosherr.WrapError(err, "Executing external CPI command: '%s'", cmdPath)
	}

	cmdOutput := CmdOutput{}
	err = json.Unmarshal([]byte(stdout), &cmdOutput)
	if err != nil {
		return CmdOutput{}, bosherr.WrapError(err, "Unmarshalling external CPI command output: STDOUT: '%s', STDERR: '%s'", stdout, stderr)
	}

	c.logger.Debug(c.logTag, cmdOutput.Log)

	if cmdOutput.Error != nil {
		return cmdOutput, bosherr.New("External CPI command for method `%s' returned an error: %s", method, cmdOutput.Error)
	}

	return cmdOutput, err
}
