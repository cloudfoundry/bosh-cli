package property

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

func Unfurl(furledProperties Map) (Map, error) {
	result := Map{}
	for compoundKey, rightValue := range furledProperties {
		compoundKeyParts := strings.Split(compoundKey, ".")

		current := result
		for _, key := range compoundKeyParts[:len(compoundKeyParts)-1] {
			newCurrent, exists := current[key]
			if !exists {
				newCurrent = Map{}
				current[key] = newCurrent
			}
			var ok bool
			current, ok = newCurrent.(Map)
			if !ok {
				return nil, bosherr.Errorf("Property structure conflict unfurling '%s' - expected map, found %#v", compoundKey, newCurrent)
			}
		}
		lastKey := compoundKeyParts[len(compoundKeyParts)-1]
		_, exists := current[lastKey]
		if exists {
			return nil, bosherr.Errorf("Property collision unfurling '%s' - multiple values specified", compoundKey)
		}
		current[lastKey] = rightValue
	}
	return result, nil
}
