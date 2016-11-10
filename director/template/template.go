package template

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

var templateFormatRegex = regexp.MustCompile(`\(\((!?[-\w\p{L}]+)\)\)`)

type Template struct {
	bytes []byte
}

type EvaluateOpts struct {
	ExpectAllKeys      bool
	UnescapedMultiline bool
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

	obj, err = t.interpolate(obj, vars, opts, missingVars)
	if err != nil {
		return []byte{}, err
	}

	if len(missingVars) > 0 {
		var missingVarKeys []string

		for v, _ := range missingVars {
			missingVarKeys = append(missingVarKeys, v)
		}

		sort.Strings(missingVarKeys)

		return []byte{}, fmt.Errorf("Expected to find variables: %s", strings.Join(missingVarKeys, ", "))
	}

	if opts.UnescapedMultiline {
		if _, ok := obj.(string); ok {
			return []byte(fmt.Sprintf("%s\n", obj)), nil
		}
	}

	bytes, err := yaml.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (t Template) interpolate(node interface{}, vars Variables, opts EvaluateOpts, missingVars map[string]struct{}) (interface{}, error) {
	switch node.(type) {
	case map[interface{}]interface{}:
		nodeMap := node.(map[interface{}]interface{})

		for k, v := range nodeMap {
			evaluatedValue, err := t.interpolate(v, vars, opts, missingVars)
			if err != nil {
				return nil, err
			}

			evaluatedKey, err := t.interpolate(k, vars, opts, missingVars)
			if err != nil {
				return nil, err
			}
			delete(nodeMap, k) // delete in case key has changed
			nodeMap[evaluatedKey] = evaluatedValue
		}
	case []interface{}:
		nodeArray := node.([]interface{})

		for i, x := range nodeArray {
			var err error
			nodeArray[i], err = t.interpolate(x, vars, opts, missingVars)
			if err != nil {
				return nil, err
			}
		}
	case string:
		nodeString := node.(string)

		for key, placeholders := range t.keysToPlaceholders(nodeString) {
			if foundVar, exists := vars[key]; exists {
				// ensure that value type is preserved when replacing the entire field
				replaceEntireField := (nodeString == fmt.Sprintf("((%s))", key) || nodeString == fmt.Sprintf("((!%s))", key))
				if replaceEntireField {
					return foundVar, nil
				}

				for _, placeholder := range placeholders {
					switch foundVar.(type) {
					case string, int, int16, int32, int64, uint, uint16, uint32, uint64:
						nodeString = strings.Replace(nodeString, placeholder, fmt.Sprintf("%v", foundVar), 1)
					default:
						return nil, fmt.Errorf("Invalid type '%T' for value '%v' and key '%s'. Supported types for interpolation within a string are integers and strings.", foundVar, foundVar, key)
					}
				}
			} else if opts.ExpectAllKeys {
				missingVars[key] = struct{}{}
			}
		}

		return nodeString, nil
	}

	return node, nil
}

func (t Template) keysToPlaceholders(value string) map[string][]string {
	matches := templateFormatRegex.FindAllSubmatch([]byte(value), -1)

	keysToPlaceholders := map[string][]string{}
	if matches == nil {
		return keysToPlaceholders
	}

	for _, match := range matches {
		capture := match[1]
		key := strings.TrimPrefix(string(capture), "!")
		if len(key) > 0 {
			keysToPlaceholders[key] = append(keysToPlaceholders[key], fmt.Sprintf("((%s))", capture))
		}
	}

	return keysToPlaceholders
}
