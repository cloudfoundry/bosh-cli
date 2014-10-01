package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

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

func (c cloud) CreateStemcell(stemcell bmstemcell.Stemcell) (cid bmstemcell.CID, err error) {
	method := "create_stemcell"
	cmdOutput, err := c.execCPICmd(method, stemcell.ImagePath, stemcell.CloudProperties)
	if err != nil {
		c.logger.Error(c.logTag, "Failed executing external CPI command with method '%s' & arguments [imagePath='%s', cloudProperties=%s]", method, stemcell.ImagePath, stemcell.CloudProperties)
		return cid, bosherr.WrapError(err, "Executing external CPI command with method `%s'", method)
	}

	// handle errors from the cpi command when exit-code = 0
	if cmdOutput.Error != nil {
		return cid, bosherr.New("External CPI command for method `%s' returned an error: %s", method, cmdOutput.Error)
	}

	// for create_stemcell, the result is a string of the stemcell cid
	cidString, ok := cmdOutput.Result.(string)
	if !ok {
		return cid, bosherr.New("Unexpected external CPI command result: '%#v'", cmdOutput.Result)
	}

	//TODO: expose CmdOutput.Error.OkToRetry?

	return bmstemcell.CID(cidString), nil
}

func (c cloud) cpiExecutablePath() string {
	return filepath.Join(c.cpiJob.JobPath, "bin", "cpi")
}

func (c cloud) execCPICmd(method string, args ...interface{}) (cmdOutput CmdOutput, err error) {
	cmdInput := CmdInput{
		Method:    method,
		Arguments: args,
		Context: CmdContext{
			DirectorUUID: c.deploymentUUID,
		},
	}
	inputBytes, err := json.Marshal(cmdInput)
	if err != nil {
		return cmdOutput, bosherr.WrapError(err, "Marshalling external CPI command input %#v", cmdInput)
	}

	inputString := string(inputBytes)
	cmdPath := c.cpiExecutablePath()
	c.logger.Debug(c.logTag, "Executing external CPI command '%s' with STDIN: '%s'", cmdPath, inputString)
	cmd := boshsys.Command{
		Name: cmdPath,
		Env: map[string]string{
			"BOSH_PACKAGES_DIR": c.cpiJob.PackagesPath,
			"BOSH_JOBS_DIR":     c.cpiJob.JobsPath,
		},
		Stdin: bytes.NewReader(inputBytes),
	}
	stdout, stderr, exitCode, err := c.cmdRunner.RunComplexCommand(cmd)
	if err != nil {
		//TODO: parse STDOUT Result.Error when exit_status != 0?
		c.logger.Error(c.logTag, "Exit Code %d when executing external CPI command '%s'\nSTDIN: '%s'\nSTDOUT: '%s'\nSTDERR: '%s'", exitCode, cmdPath, inputString, stdout, stderr)
		return cmdOutput, bosherr.WrapError(err, "Executing external CPI command: '%s'", cmdPath)
	}
	c.logger.Debug(c.logTag, "Executed external CPI command '%s' with STDOUT: '%s'", cmdPath, stdout)

	err = json.Unmarshal([]byte(stdout), &cmdOutput)
	if err != nil {
		return cmdOutput, bosherr.WrapError(err, "Unmarshalling external CPI command output: '%s'", stdout)
	}

	//TODO: write CmdOutput.Log to cpi task log?
	c.logger.Debug(c.logTag, cmdOutput.Log)

	return cmdOutput, err
}
