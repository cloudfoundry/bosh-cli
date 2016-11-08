package cmd

import (
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

// Shared
type VarFlags struct {
	VarKVs    []boshtpl.VarKV       `long:"var"      short:"v" value-name:"VAR=VALUE" description:"Set variable"`
	VarsFiles []boshtpl.VarsFileArg `long:"var-file" short:"l" value-name:"PATH"      description:"Load variables from a YAML file"`
	VarsEnvs  []boshtpl.VarsEnvArg  `long:"var-env"            value-name:"PREFIX"    description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
}

func (f VarFlags) AsVariables() boshtpl.Variables {
	vars := boshtpl.Variables{}

	for _, varsEnv := range f.VarsEnvs {
		vars = vars.Merge(varsEnv.Vars)
	}

	for _, varsFile := range f.VarsFiles {
		vars = vars.Merge(varsFile.Vars)
	}

	for _, kv := range f.VarKVs {
		vars[kv.Name] = kv.Value
	}

	return vars
}
