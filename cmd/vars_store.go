package cmd

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

type VarsStore struct {
	FS boshsys.FileSystem

	path   string
	stores map[string]boshtpl.Variables

	VarsFSStore VarsFSStore
}

var _ boshtpl.Variables = &VarsStore{}

func (s *VarsStore) IsSet() bool { return len(s.path) > 0 }

func (s *VarsStore) Get(varDef boshtpl.VariableDefinition) (interface{}, bool, error) {
	return s.store().Get(varDef)
}

func (s *VarsStore) List() ([]boshtpl.VariableDefinition, error) {
	return s.store().List()
}

func (s *VarsStore) RegisterSchemas(stores map[string]boshtpl.Variables) {
	s.stores = stores
}

func (s *VarsStore) store() boshtpl.Variables {
	const schemaSuffix = "://"

	for schema, store := range s.stores {
		if strings.HasPrefix(s.path, schema+schemaSuffix) {
			return store
		}
	}

	return s.VarsFSStore
}

func (s *VarsStore) UnmarshalFlag(data string) error {
	if len(data) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	absPath, err := s.FS.ExpandPath(data)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting absolute path '%s'", data)
	}

	(*s).path = absPath
	(*s).VarsFSStore = NewVarsFSStore(s.FS, absPath)

	return nil
}
