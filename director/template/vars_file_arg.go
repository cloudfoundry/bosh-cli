package template

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type VarsFileArg struct {
	FS boshsys.FileSystem

	Vars StaticVariables
}

func (a *VarsFileArg) UnmarshalFlag(filePath string) error {
	if len(filePath) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	fileScope := ""
	fileSplit := strings.SplitN(filePath, "=", 2)

	if len(fileSplit) == 2 {
		fileScope = fileSplit[0]
		filePath = fileSplit[1]
	}

	bytes, err := a.FS.ReadFile(filePath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Reading variables file '%s'", filePath)
	}

	var vars StaticVariables

	err = yaml.Unmarshal(bytes, &vars)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deserializing variables file '%s'", filePath)
	}

	if fileScope != "" {
		scopedVars := map[interface{}]interface{}{}

		for k, v := range vars {
			scopedVars[k] = v
		}

		vars = StaticVariables{
			fileScope: scopedVars,
		}
	}

	(*a).Vars = vars

	return nil
}
