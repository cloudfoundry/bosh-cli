package opts

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CurlHeader struct {
	Name  string
	Value string
}

func (a *CurlHeader) UnmarshalFlag(data string) error {
	pieces := strings.SplitN(data, ": ", 2)
	if len(pieces) != 2 {
		return bosherr.Errorf("Expected header '%s' to be in format 'name: value'", data)
	}

	if len(pieces[0]) == 0 {
		return bosherr.Errorf("Expected header '%s' to specify non-empty name", data)
	}

	if len(pieces[1]) == 0 {
		return bosherr.Errorf("Expected header '%s' to specify non-empty value", data)
	}

	*a = CurlHeader{Name: pieces[0], Value: pieces[1]}

	return nil
}
