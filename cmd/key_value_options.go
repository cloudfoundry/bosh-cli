package cmd

import (
	"regexp"
	"strconv"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

var (
	integerOptionValue = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	floatOptionValue   = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+([eE][+-]?[0-9]+)?|[eE][+-]?[0-9]+)$`)
)

func parseKeyValueOptions(values []string) (map[string]interface{}, error) {
	if len(values) == 0 {
		return nil, nil
	}

	options := map[string]interface{}{}

	for _, value := range values {
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 || len(parts[0]) == 0 {
			return nil, bosherr.Errorf("Expected KEY=VALUE option, got '%s'", value)
		}

		options[parts[0]] = parseOptionValue(parts[1])
	}

	return options, nil
}

// parseOptionValue gives unquoted CLI values the same scalar types they would
// have in YAML (booleans and numbers), so blobstore providers receive JSON
// values with the types they expect. Quote a value to force it to stay a
// string.
func parseOptionValue(value string) interface{} {
	if len(value) >= 2 {
		if value[0] == '"' && value[len(value)-1] == '"' {
			return value[1 : len(value)-1]
		}

		if value[0] == '\'' && value[len(value)-1] == '\'' {
			return value[1 : len(value)-1]
		}
	}

	switch value {
	case "true":
		return true
	case "false":
		return false
	}

	if integerOptionValue.MatchString(value) {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}

	if floatOptionValue.MatchString(value) {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}

	return value
}
