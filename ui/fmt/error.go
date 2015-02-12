package fmt

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

func MultilineError(err error) string {
	return prefixingMultilineError(err, "")
}

func prefixingMultilineError(err error, prefix string) string {
	compoundErr, ok := err.(bosherr.ComplexError)
	if ok {
		return prefix + compoundErr.Err.Error() + ":\n" + prefixingMultilineError(compoundErr.Cause, prefix+"  ")
	}
	multiErr, ok := err.(bosherr.MultiError)
	if ok {
		lines := make([]string, len(multiErr.Errors), len(multiErr.Errors))
		for i, sibling := range multiErr.Errors {
			lines[i] = prefixingMultilineError(sibling, prefix)
		}
		return strings.Join(lines, "\n")
	}
	return prefix + err.Error()
}
