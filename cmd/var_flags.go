package cmd

import (
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

// Shared
type VarFlags struct {
	VarKVs    []boshtpl.VarKV       `long:"var"       short:"v" value-name:"[SECRET=KEY]" description:"Variable flag that can be used for filling in template values in configuration"`
	VarsFiles []boshtpl.VarsFileArg `long:"var-files" short:"l"                           description:"Variable flag that can be used for filling in template values in configuration from a YAML file"`
	VarsEnvs  []boshtpl.VarsEnvArg  `long:"var-env"             value-name:"PREFIX"       description:"Variable flag that can be used for filling in template values in configuration from environment variables"`
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
