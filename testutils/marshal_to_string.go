package testutils

import (
	"gopkg.in/yaml.v2"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

func MarshalToString(input interface{}) (string, error) {
	bytes, err := yaml.Marshal(input)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Marshaling to string: %#v", input)
	}

	return string(bytes), nil
}
