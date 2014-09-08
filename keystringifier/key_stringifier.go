package keystringifier

import (
	"reflect"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type KeyStringifier interface {
	ConvertMap(m map[interface{}]interface{}) (map[string]interface{}, error)
}

type keyStringifier struct{}

func NewKeyStringifier() keyStringifier { return keyStringifier{} }

// convertMapToStringKeyMap converts map to string keyed map.
func (sk keyStringifier) ConvertMap(m map[interface{}]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	for name, val := range m {
		nameStr, ok := name.(string)
		if !ok {
			return result, bosherr.New("Map contains non-string key %v", name)
		}

		convertedVal, err := sk.convertInterface(val)
		if err != nil {
			return result, err
		}

		result[nameStr] = convertedVal
	}

	return result, nil
}

func (sk keyStringifier) convertInterface(val interface{}) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.Map:
		valMap, ok := val.(map[interface{}]interface{})
		if !ok {
			return nil, bosherr.New("Converting map %v", val)
		}

		return sk.ConvertMap(valMap)

	case reflect.Slice:
		valSlice, ok := val.([]interface{})
		if !ok {
			return nil, bosherr.New("Converting slice %v", val)
		}

		slice := make([]interface{}, len(valSlice))

		for i, v := range valSlice {
			convertedVal, err := sk.convertInterface(v)
			if err != nil {
				return nil, err
			}

			slice[i] = convertedVal
		}

		return slice, nil

	default:
		return val, nil
	}
}
