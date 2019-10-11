package opts

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/go-patch/patch"
	"gopkg.in/yaml.v2"
)

type OpsFileArg struct {
	FS boshsys.FileSystem

	Ops patch.Ops
}

func (a *OpsFileArg) UnmarshalFlag(filePath string) error {
	if a.FS == nil {
		a.FS = boshsys.NewOsFileSystemWithStrictTempRoot(boshlog.NewLogger(boshlog.LevelNone))
	}

	if len(filePath) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	bytes, err := a.FS.ReadFile(filePath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Reading ops file '%s'", filePath)
	}

	var opDefs []patch.OpDefinition

	err = yaml.Unmarshal(bytes, &opDefs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deserializing ops file '%s'", filePath)
	}

	for i := range opDefs {
		errorStr := fmt.Sprintf("operation [%d] in %s failed", i, filePath)
		opDefs[i].Error = &errorStr
	}

	ops, err := patch.NewOpsFromDefinitions(opDefs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Building ops")
	}

	(*a).Ops = ops

	return nil
}
