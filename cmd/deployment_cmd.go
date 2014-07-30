package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

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
}

func NewDeploymentCmd(ui bmui.UI, boshMicroPath string) *deploymentCmd {
	return &deploymentCmd{
		ui:            ui,
		boshMicroPath: boshMicroPath,
	}
}

func (c *deploymentCmd) Run(args []string) error {
	if args == nil || len(args) < 1 {
		deploymentJson, err := c.readBoshMicroFile()
		if err != nil || deploymentJson.Deployment == "" {
			c.ui.Error("Deployment not set")
			return errors.New("Deployment not set")
		} else {
			c.ui.Say(fmt.Sprintf("Current deployment is '%s'", deploymentJson.Deployment))
			return nil
		}
	}

	manifestFilePath := args[0]
	if _, err := os.Stat(manifestFilePath); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Deployment command manifest path %s does not exist", manifestFilePath))
	}

	boshMicroPath := path.Join(c.boshMicroPath, BOSH_MICRO_FILENAME)

	jsonContentStruct := &deploymentFileJson{Deployment: manifestFilePath}
	jsonContent, err := json.MarshalIndent(jsonContentStruct, "", "  ")
	if err != nil {
		return errors.New("Could not marshal JSON content %s")
	}

	err = ioutil.WriteFile(boshMicroPath, jsonContent, os.ModePerm)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not write to file %s", boshMicroPath))
	}

	c.ui.Say(fmt.Sprintf("Deployment set to '%s'", manifestFilePath))
	return nil
}

func (c *deploymentCmd) readBoshMicroFile() (*deploymentFileJson, error) {
	boshMicroPath := path.Join(c.boshMicroPath, BOSH_MICRO_FILENAME)
	content, err := ioutil.ReadFile(boshMicroPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not read BOSH micro file %s", boshMicroPath))
	}

	jsonContentStruct := &deploymentFileJson{}
	err = json.Unmarshal(content, jsonContentStruct)
	if err != nil {
		return nil, errors.New("Could not marshal JSON content %s")
	}

	return jsonContentStruct, nil
}
