package completion

import (
	"fmt"

	"github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type BlindUI struct{}

var errNotSupported = fmt.Errorf("function not supported")

func (ui *BlindUI) ErrorLinef(_ string, _ ...interface{}) {}

func (ui *BlindUI) PrintLinef(_ string, _ ...interface{}) {}

func (ui *BlindUI) BeginLinef(_ string, _ ...interface{}) {
}

func (ui *BlindUI) EndLinef(_ string, _ ...interface{}) {
}

func (ui *BlindUI) PrintBlock(_ []byte) {
}

func (ui *BlindUI) PrintErrorBlock(_ string) {
}

func (ui *BlindUI) PrintTable(_ table.Table) {
}

func (ui *BlindUI) PrintTableFiltered(_ table.Table, _ []table.Header) {
}

func (ui *BlindUI) AskForText(_ string) (string, error) {
	return "", errNotSupported
}

func (ui *BlindUI) AskForTextWithDefaultValue(_, _ string) (string, error) {
	return "", errNotSupported
}

func (ui *BlindUI) AskForChoice(_ string, _ []string) (int, error) {
	return 0, errNotSupported
}

func (ui *BlindUI) AskForPassword(_ string) (string, error) {
	return "", errNotSupported
}

func (ui *BlindUI) AskForConfirmation() error {
	return errNotSupported
}

func (ui *BlindUI) AskForConfirmationWithLabel(_ string) error {
	return errNotSupported
}

func (ui *BlindUI) IsInteractive() bool {
	return false
}

func (ui *BlindUI) Flush() {}
