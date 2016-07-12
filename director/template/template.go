package template

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-multierror"
)

var templateFormatRegex = regexp.MustCompile(`\(\(([-\w\p{L}]+)\)\)`)

type Template struct {
	bytes []byte
}

func NewTemplate(bytes []byte) Template {
	return Template{bytes: bytes}
}

func (t Template) Evaluate(vars Variables) ([]byte, error) {
	var errs error

	replaceFunc := func(match []byte) []byte {
		key := string(templateFormatRegex.FindSubmatch(match)[1])

		value, found := vars[key]
		if !found {
			err := fmt.Errorf("unbound variable in template: '%s'", key)
			errs = multierror.Append(errs, err)
			return match
		}

		saveValue, err := json.Marshal(value)
		if err != nil {
			panic("Unexpected marshaling error in Template")
		}

		return []byte(saveValue)
	}

	return templateFormatRegex.ReplaceAllFunc(t.bytes, replaceFunc), errs
}
