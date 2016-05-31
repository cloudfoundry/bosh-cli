package template

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type VarKV struct {
	Name  string
	Value string
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

	*a = VarKV{Name: pieces[0], Value: pieces[1]}

	return nil
}
