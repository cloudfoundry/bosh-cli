package testutils

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

func MarshalToString(input interface{}) (string, error) {
	bytes, err := candiedyaml.Marshal(input)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Marshaling to string: %#v", input)
	}

	return string(bytes), nil
}
