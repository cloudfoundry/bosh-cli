package template

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type VarsFileArg struct {
	FS boshsys.FileSystem

	Vars Variables
}

func (a *VarsFileArg) UnmarshalFlag(data string) error {
	if len(data) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	bytes, err := a.FS.ReadFile(data)
	if err != nil {
		return bosherr.WrapErrorf(err, "Reading variables file '%s'", data)
	}

	var vars Variables

	err = yaml.Unmarshal(bytes, &vars)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deserializing variables file '%s'", data)
	}

	(*a).Vars = vars

	return nil
}
