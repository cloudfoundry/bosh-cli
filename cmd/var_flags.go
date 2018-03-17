package cmd

import (
	cfgtypes "github.com/cloudfoundry/config-server/types"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

// Shared
type VarFlags struct {
	VarKVs    []boshtpl.VarKV       `long:"var"        short:"v" value-name:"VAR=VALUE" description:"Set variable"`
	VarFiles  []boshtpl.VarFileArg  `long:"var-file"             value-name:"VAR=PATH"  description:"Set variable to file contents"`
	VarsFiles []boshtpl.VarsFileArg `long:"vars-file"  short:"l" value-name:"PATH"      description:"Load variables from a YAML file"`
	VarsEnvs  []boshtpl.VarsEnvArg  `long:"vars-env"             value-name:"PREFIX"    description:"Load variables from environment variables (e.g.: 'MY' to load MY_var=value)"`
	VarsStore *VarsStore            `long:"vars-store"           value-name:"PATH"      description:"Load/save variables from/to a YAML file"`
}

func (f VarFlags) AsVariables() boshtpl.Variables {
	var firstToUse []boshtpl.Variables

	staticVars := boshtpl.StaticVariables{}

	for i, _ := range f.VarsEnvs {
		for k, v := range f.VarsEnvs[i].Vars {
			staticVars[k] = v
		}
	}

	for i, _ := range f.VarsFiles {
		for k, v := range f.VarsFiles[i].Vars {
			staticVars[k] = v
		}
	}

	for i, _ := range f.VarFiles {
		for k, v := range f.VarFiles[i].Vars {
			staticVars[k] = v
		}
	}

	for _, kv := range f.VarKVs {
		staticVars[kv.Name] = kv.Value
	}

	firstToUse = append(firstToUse, staticVars)

	if f.VarsStore != nil && f.VarsStore.IsSet() {
		firstToUse = append(firstToUse, f.VarsStore)
	}

	vars := boshtpl.NewMultiVars(firstToUse)

	if f.VarsStore != nil && f.VarsStore.VarsFSStore.IsSet() {
		f.VarsStore.VarsFSStore.ValueGeneratorFactory = cfgtypes.NewValueGeneratorConcrete(NewVarsCertLoader(vars))
	}

	return vars
}
