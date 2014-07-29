package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

const (
	BOSH_MICRO_FILENAME = ".bosh_micro"
)

type deploymentCmd struct {
	args []string
}

func NewDeploymentCmd() *deploymentCmd {
	return &deploymentCmd{}
}

func (f *deploymentCmd) Run(args []string) error {
	f.args = args

	if f.args == nil {
		return errors.New("Deployment command argument cannot be nil")
	}

	if len(f.args) < 1 {
		return errors.New("Deployment command arguments must have at least one arg")
	}

	manifestFilePath := f.args[0]
	if _, err := os.Stat(manifestFilePath); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Deployment command manifest path %s does not exist", manifestFilePath))
	}

	usr, err := user.Current()
	if err != nil {
		return errors.New("Could not access current user")
	}

	boshMicroPath := path.Join(usr.HomeDir, BOSH_MICRO_FILENAME)
	err = ioutil.WriteFile(boshMicroPath, []byte(manifestFilePath), os.ModePerm)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not write to file %s", boshMicroPath))
	}

	return nil
}
