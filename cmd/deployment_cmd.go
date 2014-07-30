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
	if args == nil {
		return errors.New("Deployment command argument cannot be nil")
	}

	if len(args) < 1 {
		return errors.New("Deployment command arguments must have at least one arg")
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

	c.ui.Say(fmt.Sprintf("Deployment set to `%s'", manifestFilePath))
	return nil
}
