package property

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

// Merge copies values from the right map into the left map, recursively
func Merge(left Map, right Map) error {
	for key, rightValue := range right {
		leftValue, exists := left[key]
		if exists {
			leftValueMap, leftIsMap := leftValue.(Map)
			rightValueMap, rightIsMap := rightValue.(Map)
			if leftIsMap != rightIsMap {
				return bosherr.Errorf("Property collision merging '%s'", key)
			}
			if leftIsMap && rightIsMap {
				Merge(leftValueMap, rightValueMap)
			} else {
				left[key] = rightValue
			}
		} else {
			left[key] = rightValue
		}
	}
	return nil
}
