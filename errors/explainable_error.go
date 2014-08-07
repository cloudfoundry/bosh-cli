package errors

import (
	"fmt"
)

type ExplainableError struct {
	errors []error
}

func NewExplainableError(errors []error) error {
	return ExplainableError{errors: errors}
}

func (e ExplainableError) Error() string {
	output := ""

	for _, err := range e.errors {
		if output != "" {
			output += "\n"
		}

		output += fmt.Sprintf("%s", err.Error())
	}

	return output
}
