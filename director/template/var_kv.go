package template

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
	"fmt"
)

type VarKV struct {
	Name  string
	Value interface{}
}

func (a *VarKV) UnmarshalFlag(data string) error {
	pieces := strings.Split(data, "=")
	if len(pieces) != 2 {
		return bosherr.Errorf(
			"Expected var '%s' to be in format 'name=value'", data)
	}

	if len(pieces[0]) == 0 {
		return bosherr.Errorf(
			"Expected var '%s' to specify non-empty name", data)
	}

	if len(pieces[1]) == 0 {
		return bosherr.Errorf(
			"Expected var '%s' to specify non-empty value", data)
	}

	var vars map[string]interface{}
	err := yaml.Unmarshal([]byte(fmt.Sprintf("{%v: %v}", pieces[0], pieces[1])), &vars)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deserializing variables '%s'", data)
	}
	*a = VarKV{Name: pieces[0], Value: vars[pieces[0]]}

	return nil
}
