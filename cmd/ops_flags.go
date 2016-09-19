package cmd

import (
	"github.com/cppforlife/go-patch"
)

// Shared
type OpsFlags struct {
	OpsFiles []OpsFileArg `long:"ops-file" short:"o" description:"Path to a YAML file that contains list of operations to modify template"`
}

func (f OpsFlags) AsOps() patch.Ops {
	var ops patch.Ops

	for _, opsFile := range f.OpsFiles {
		ops = append(ops, opsFile.Ops...)
	}

	return ops
}
