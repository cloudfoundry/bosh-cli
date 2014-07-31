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
	BoshMicroFilename = ".bosh_micro.json"
)

type deploymentFileJSON struct {
	Deployment string `json:"deployment"`
}

type deploymentCmd struct {
	ui            bmui.UI
	boshMicroPath string
	fs            boshsys.FileSystem
}

func NewDeploymentCmd(ui bmui.UI, boshMicroPath string, fs boshsys.FileSystem) *deploymentCmd {
	fullBoshMicroPath := path.Join(boshMicroPath, BoshMicroFilename)
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
	deploymentJSON, err := c.readBoshMicroFile()

	if err != nil || deploymentJSON.Deployment == "" {
		c.ui.Error("Deployment not set")
		return errors.New("Deployment not set")
	}

	c.ui.Say(fmt.Sprintf("Current deployment is '%s'", deploymentJSON.Deployment))
	return nil
}

func (c *deploymentCmd) setDeployment(manifestFilePath string) error {
	if !c.fs.FileExists(manifestFilePath) {
		return fmt.Errorf("Deployment command manifest path %s does not exist", manifestFilePath)
	}

	var err error
	var deploymentJSON *deploymentFileJSON
	if !c.fs.FileExists(c.boshMicroPath) {
		deploymentJSON = &deploymentFileJSON{}
	} else {
		deploymentJSON, err = c.readBoshMicroFile()
		if err != nil {
			return err
		}
	}

	deploymentJSON.Deployment = manifestFilePath

	err = c.saveBoshMicroFile(manifestFilePath, deploymentJSON)
	if err != nil {
		return err
	}

	c.ui.Say(fmt.Sprintf("Deployment set to '%s'", manifestFilePath))
	return nil
}

func (c *deploymentCmd) saveBoshMicroFile(manifestFilePath string, deploymentJSON *deploymentFileJSON) error {
	jsonContent, err := json.MarshalIndent(deploymentJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("Could not marshal JSON content '%s'", manifestFilePath)
	}

	err = c.fs.WriteFile(c.boshMicroPath, jsonContent)
	if err != nil {
		return fmt.Errorf("Could not write to BOSH micro file %s", c.boshMicroPath)
	}

	return nil
}

func (c *deploymentCmd) readBoshMicroFile() (*deploymentFileJSON, error) {
	content, err := c.fs.ReadFile(c.boshMicroPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read BOSH micro file %s", c.boshMicroPath)
	}

	jsonContentStruct := &deploymentFileJSON{}
	err = json.Unmarshal(content, jsonContentStruct)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal JSON content '%s'", c.boshMicroPath)
	}

	return jsonContentStruct, nil
}
