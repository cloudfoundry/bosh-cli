package cmd

import (
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
)

func (f VarFlags) AsVariables() boshtpl.Variables {
	vars := boshtpl.Variables{}

	for _, varsFile := range f.VarsFiles {
		vars = vars.Merge(varsFile.Vars)
	}

	for _, kv := range f.VarKVs {
		vars[kv.Name] = kv.Value
	}

	return vars
}
