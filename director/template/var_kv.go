package template

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
)

type VarKV struct {
	Name  string
	Value interface{}
}

func trimWrappingQuotes(s string) string {
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		s = s[1 : len(s)-1]
		s = trimWrappingQuotes(s)
	}
	return s
}
func (a *VarKV) UnmarshalFlag(data string) error {
	pieces := strings.SplitN(data, "=", 2)
	const nameIndex, valueIndex = 0, 1
	if len(pieces) != 2 {
		return bosherr.Errorf("Expected var '%s' to be in format 'name=value'", data)
	}

	if len(pieces[nameIndex]) == 0 {
		return bosherr.Errorf("Expected var '%s' to specify non-empty name", data)
	}

	if len(pieces[valueIndex]) == 0 {
		return bosherr.Errorf("Expected var '%s' to specify non-empty value", data)
	}

	var vars interface{}

	err := yaml.Unmarshal([]byte(pieces[valueIndex]), &vars)
	// If for whatever reason YAML unmarshal fails, treat the second part of pieces as a string
	if err != nil {
		vars = pieces[valueIndex]
	}
	//yaml.Unmarshal returns a string if the input is not valid yaml.
	if _, ok := vars.(string); ok {
		//in that case, we return the initial flag value as the Unmarshal
		//process strips newlines. Stripping the quotes is required to keep
		//backwards compability (YAML unmarshal also removed wrapping quotes from the value).
		*a = VarKV{Name: pieces[nameIndex], Value: trimWrappingQuotes(pieces[valueIndex])}
	} else {
		//otherwise, return the parsed YAML object
		*a = VarKV{Name: pieces[nameIndex], Value: vars}
	}

	return nil
}
