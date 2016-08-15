package template

import (
	"gopkg.in/yaml.v2"
	"regexp"
)

var templateFormatRegex = regexp.MustCompile(`^\(\(([-\w\p{L}]+)\)\)$`)

type Template struct {
	bytes []byte
}

func NewTemplate(bytes []byte) Template {
	return Template{bytes: bytes}
}

func (t Template) Evaluate(vars Variables) ([]byte, error) {
	var templateYaml interface{}

	err := yaml.Unmarshal(t.bytes, &templateYaml)
	if err != nil {
		return []byte{}, err
	}

	compiledTemplate := t.interpolate(templateYaml, vars)

	bytes, err := yaml.Marshal(compiledTemplate)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

func (t Template) interpolate(node interface{}, vars Variables) interface{} {
	switch node.(type) {
	case map[interface{}]interface{}:
		nodeMap := node.(map[interface{}]interface{})
		for k, v := range nodeMap {
			evaluatedValue := t.interpolate(v, vars)
			keyAsString, ok := k.(string)
			if ok {
				newKey, ok := t.needsEvaluation(keyAsString)
				if ok {
					foundVarKey, exists := vars[newKey]
					if exists {
						delete(nodeMap, k)
						k = foundVarKey
					}
				}
			}

			nodeMap[k] = evaluatedValue
		}
	case []interface{}:
		nodeArray := node.([]interface{})
		for i, x := range nodeArray {
			nodeArray[i] = t.interpolate(x, vars)
		}
	case string:
		key, found := t.needsEvaluation(node.(string))
		if found {
			foundVar, exists := vars[key]
			if exists {
				return foundVar
			}
		}
	default:
	}

	return node
}

func (t Template) needsEvaluation(value string) (string, bool) {
	found := templateFormatRegex.FindAllSubmatch([]byte(value), 1)
	if len(found) != 0 && len(found[0]) != 0 {
		return string(found[0][1]), true
	}
	return "", false
}
