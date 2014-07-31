package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

const (
	BOSH_MICRO_FILENAME = ".bosh_micro.json"
)

type deploymentFileJson struct {
	Deployment string `json:"deployment"`
}

type deploymentCmd struct {
	ui            bmui.UI
	boshMicroPath string
	fs            boshsys.FileSystem
}

func NewDeploymentCmd(ui bmui.UI, boshMicroPath string, fs boshsys.FileSystem) *deploymentCmd {
	fullBoshMicroPath := path.Join(boshMicroPath, BOSH_MICRO_FILENAME)
	return &deploymentCmd{
		ui:            ui,
		boshMicroPath: fullBoshMicroPath,
		fs:            fs,
	}
}

func (c *deploymentCmd) Run(args []string) error {
	if args == nil || len(args) < 1 {
		return c.showDeploymentStatus()
	}

	manifestFilePath := args[0]
	return c.setDeployment(manifestFilePath)
}

func (c *deploymentCmd) showDeploymentStatus() error {
	deploymentJson, err := c.readBoshMicroFile()

	if err != nil || deploymentJson.Deployment == "" {
		c.ui.Error("Deployment not set")
		return errors.New("Deployment not set")
	} else {
		c.ui.Say(fmt.Sprintf("Current deployment is '%s'", deploymentJson.Deployment))
		return nil
	}
}

func (c *deploymentCmd) setDeployment(manifestFilePath string) error {
	if !c.fs.FileExists(manifestFilePath) {
		return errors.New(fmt.Sprintf("Deployment command manifest path %s does not exist", manifestFilePath))
	}

	var err error
	var deploymentJson *deploymentFileJson
	if !c.fs.FileExists(c.boshMicroPath) {
		deploymentJson = &deploymentFileJson{}
	} else {
		deploymentJson, err = c.readBoshMicroFile()
		if err != nil {
			return err
		}
	}

	deploymentJson.Deployment = manifestFilePath

	err = c.saveBoshMicroFile(manifestFilePath, deploymentJson)
	if err != nil {
		return err
	}

	c.ui.Say(fmt.Sprintf("Deployment set to '%s'", manifestFilePath))
	return nil
}

func (c *deploymentCmd) saveBoshMicroFile(manifestFilePath string, deploymentJson *deploymentFileJson) error {
	jsonContent, err := json.MarshalIndent(deploymentJson, "", "  ")
	if err != nil {
		return errors.New(fmt.Sprintf("Could not marshal JSON content '%s'", manifestFilePath))
	}

	err = c.fs.WriteFile(c.boshMicroPath, jsonContent)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not write to BOSH micro file %s", c.boshMicroPath))
	}

	return nil
}

func (c *deploymentCmd) readBoshMicroFile() (*deploymentFileJson, error) {
	content, err := c.fs.ReadFile(c.boshMicroPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not read BOSH micro file %s", c.boshMicroPath))
	}

	jsonContentStruct := &deploymentFileJson{}
	err = json.Unmarshal(content, jsonContentStruct)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not unmarshal JSON content '%s'", c.boshMicroPath))
	}

	return jsonContentStruct, nil
}
