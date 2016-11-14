package template

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

var (
	templateFormatRegex         = regexp.MustCompile(`\(\((!?[-\w\p{L}]+)\)\)`)
	templateFormatAnchoredRegex = regexp.MustCompile("\\A" + templateFormatRegex.String() + "\\z")
)

type Template struct {
	bytes []byte
}

type EvaluateOpts struct {
	ExpectAllKeys         bool
	PostVarSubstitutionOp patch.Op
	UnescapedMultiline    bool
}

func NewTemplate(bytes []byte) Template {
	return Template{bytes: bytes}
}

func (t Template) Evaluate(vars Variables, op patch.Op, opts EvaluateOpts) ([]byte, error) {
	var obj interface{}

	err := yaml.Unmarshal(t.bytes, &obj)
	if err != nil {
		return []byte{}, err
	}

	if op != nil {
		obj, err = op.Apply(obj)
		if err != nil {
			return []byte{}, err
		}
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

	if opts.PostVarSubstitutionOp != nil {
		obj, err = opts.PostVarSubstitutionOp.Apply(obj)
		if err != nil {
			return []byte{}, err
		}
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
	switch typedNode := node.(type) {
	case map[interface{}]interface{}:
		for k, v := range typedNode {
			evaluatedValue, err := t.interpolate(v, vars, opts, missingVars)
			if err != nil {
				return nil, err
			}

			evaluatedKey, err := t.interpolate(k, vars, opts, missingVars)
			if err != nil {
				return nil, err
			}

			delete(typedNode, k) // delete in case key has changed
			typedNode[evaluatedKey] = evaluatedValue
		}

	case []interface{}:
		for i, x := range typedNode {
			var err error
			typedNode[i], err = t.interpolate(x, vars, opts, missingVars)
			if err != nil {
				return nil, err
			}
		}

	case string:
		for _, key := range t.keys(typedNode) {
			if foundVar, exists := vars[key]; exists {
				// ensure that value type is preserved when replacing the entire field
				if templateFormatAnchoredRegex.MatchString(typedNode) {
					return foundVar, nil
				}

				switch foundVar.(type) {
				case string, int, int16, int32, int64, uint, uint16, uint32, uint64:
					foundVarStr := fmt.Sprintf("%v", foundVar)
					typedNode = strings.Replace(typedNode, fmt.Sprintf("((%s))", key), foundVarStr, -1)
					typedNode = strings.Replace(typedNode, fmt.Sprintf("((!%s))", key), foundVarStr, -1)
				default:
					errMsg := "Invalid type '%T' for value '%v' and key '%s'. Supported types for interpolation within a string are integers and strings."
					return nil, fmt.Errorf(errMsg, foundVar, foundVar, key)
				}
			} else if opts.ExpectAllKeys {
				missingVars[key] = struct{}{}
			}
		}

		return typedNode, nil
	}

	return node, nil
}

func (t Template) keys(value string) []string {
	var keys []string

	for _, match := range templateFormatRegex.FindAllSubmatch([]byte(value), -1) {
		keys = append(keys, strings.TrimPrefix(string(match[1]), "!"))
	}

	return keys
}
