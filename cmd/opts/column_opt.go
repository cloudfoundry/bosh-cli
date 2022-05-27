package opts

import (
	"github.com/cloudfoundry/bosh-cli/v6/ui/table"
)

type ColumnOpt struct {
	table.Header
}

func (a *ColumnOpt) UnmarshalFlag(arg string) error {
	a.Key = table.KeyifyHeader(arg)
	a.Hidden = false

	return nil
}
