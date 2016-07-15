package template

import (
	"gopkg.in/yaml.v2"

	"fmt"
	"regexp"

	"github.com/hashicorp/go-multierror"
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
		return nil, err
	}

	compiledTemplate, err := evaluate(templateYaml, vars)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(compiledTemplate)
}

func evaluate(node interface{}, vars Variables) (interface{}, error) {
	var errs error
	var err error
	switch node.(type) {
	case map[interface{}]interface{}:
		nodeMap := node.(map[interface{}]interface{})
		for k, v := range nodeMap {
			evaluatedValue, err := evaluate(v, vars)
			if err != nil {
				errs = multierror.Append(errs, err)
			}

			newKey, ok := needsEvaluation(fmt.Sprintf("%v", k))
			if ok {
				foundVarKey, exists := vars[newKey]
				if exists {
					delete(nodeMap, k)
					k = foundVarKey
				}
			}
			nodeMap[k] = evaluatedValue
		}
	case []interface{}:
		nodeArray := node.([]interface{})
		for i, x := range nodeArray {
			nodeArray[i], err = evaluate(x, vars)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	case string:
		key, found := needsEvaluation(node.(string))
		if found {
			foundVar, exists := vars[key]
			if exists {
				return foundVar, nil
			}
		}
	default:
	}

	return node, errs
}

func needsEvaluation(str string) (string, bool) {
	found := templateFormatRegex.FindAllSubmatch([]byte(str), 1)
	if len(found) != 0 && len(found[0]) != 0 {
		return string(found[0][1]), true
	}
	return "", false
}
