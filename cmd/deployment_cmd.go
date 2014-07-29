package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const (
	BOSH_MICRO_FILENAME = ".bosh_micro.json"
)

type deploymentFileJson struct {
	Deployment string `json:"deployment"`
}

type deploymentCmd struct {
	boshMicroPath string
}

func NewDeploymentCmd(boshMicroPath string) *deploymentCmd {
	return &deploymentCmd{boshMicroPath: boshMicroPath}
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

	return nil
}
