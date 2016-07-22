package cmd

import (
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
)


// Shared
type VarFlags struct {
	VarKVs    []boshtpl.VarKV       `long:"var"       short:"v" value-name:"[SECRET=KEY]" description:"Variable flag that can be used for filling in template values in configuration"`
	VarsFiles []boshtpl.VarsFileArg `long:"var-files" short:"l"                           description:"Variable flag that can be used for filling in template values in configuration from a YAML file"`
}

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
