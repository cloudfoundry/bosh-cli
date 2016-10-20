package template

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

var templateFormatRegex = regexp.MustCompile(`^\(\(([-\w\p{L}]+)\)\)$`)

type Template struct {
	bytes []byte
}

type EvaluateOpts struct {
	ExpectAllKeys bool
}

func NewTemplate(bytes []byte) Template {
	return Template{bytes: bytes}
}

func (t Template) Evaluate(vars Variables, ops patch.Ops, opts EvaluateOpts) ([]byte, error) {
	var obj interface{}

	err := yaml.Unmarshal(t.bytes, &obj)
	if err != nil {
		return []byte{}, err
	}

	obj, err = ops.Apply(obj)
	if err != nil {
		return []byte{}, err
	}

	missingVars := map[string]struct{}{}

	obj = t.interpolate(obj, vars, opts, missingVars)

	if len(missingVars) > 0 {
		var missingVarKeys []string

		for v, _ := range missingVars {
			missingVarKeys = append(missingVarKeys, v)
		}

		sort.Strings(missingVarKeys)

		return []byte{}, fmt.Errorf("Expected to find variables: %s", strings.Join(missingVarKeys, ", "))
	}

	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (t Template) interpolate(node interface{}, vars Variables, opts EvaluateOpts, missingVars map[string]struct{}) interface{} {
	switch node.(type) {
	case map[interface{}]interface{}:
		nodeMap := node.(map[interface{}]interface{})

		for k, v := range nodeMap {
			evaluatedValue := t.interpolate(v, vars, opts, missingVars)

			if keyAsString, ok := k.(string); ok {
				if key, eval := t.needsEvaluation(keyAsString); eval {
					if foundVarKey, exists := vars[key]; exists {
						delete(nodeMap, k)
						k = foundVarKey
					} else if opts.ExpectAllKeys {
						missingVars[key] = struct{}{}
					}
				}
			}

			nodeMap[k] = evaluatedValue
		}

	case []interface{}:
		nodeArray := node.([]interface{})

		for i, x := range nodeArray {
			nodeArray[i] = t.interpolate(x, vars, opts, missingVars)
		}

	case string:
		if key, found := t.needsEvaluation(node.(string)); found {
			if foundVar, exists := vars[key]; exists {
				return foundVar
			} else if opts.ExpectAllKeys {
				missingVars[key] = struct{}{}
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
