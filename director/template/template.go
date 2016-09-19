package template

import (
	"regexp"

	"github.com/cppforlife/go-patch"
	"gopkg.in/yaml.v2"
)

var templateFormatRegex = regexp.MustCompile(`^\(\(([-\w\p{L}]+)\)\)$`)

type Template struct {
	bytes []byte
}

func NewTemplate(bytes []byte) Template {
	return Template{bytes: bytes}
}

func (t Template) Evaluate(vars Variables, ops patch.Ops) ([]byte, error) {
	var obj interface{}

	err := yaml.Unmarshal(t.bytes, &obj)
	if err != nil {
		return []byte{}, err
	}

	obj, err = ops.Apply(obj)
	if err != nil {
		return []byte{}, err
	}

	obj = t.interpolate(obj, vars)

	bytes, err := yaml.Marshal(obj)
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

			if keyAsString, ok := k.(string); ok {
				if newKey, eval := t.needsEvaluation(keyAsString); eval {
					if foundVarKey, exists := vars[newKey]; exists {
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
		if key, found := t.needsEvaluation(node.(string)); found {
			if foundVar, exists := vars[key]; exists {
				return foundVar
			}
		}
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
