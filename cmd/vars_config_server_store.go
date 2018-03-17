package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bconfigserver "github.com/cloudfoundry/bosh-cli/configserver"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

type VarsConfigServerStore struct {
	client bconfigserver.Client
}

var _ boshtpl.Variables = VarsConfigServerStore{}

func NewConfigServerVarsStore(client bconfigserver.Client) *VarsConfigServerStore {
	return &VarsConfigServerStore{client}
}

func (s VarsConfigServerStore) Get(varDef boshtpl.VariableDefinition) (interface{}, bool, error) {
	found, err := s.client.Exists(varDef.Name)
	if err != nil {
		return nil, false, err
	}

	if found {
		content, err := s.client.Read(varDef.Name)
		if err != nil {
			return nil, false, err
		}

		return content, true, nil
	}

	if len(varDef.Type) == 0 {
		return nil, false, nil
	}

	val, err := s.client.Generate(varDef.Name, varDef.Type, varDef.Options)
	if err != nil {
		return nil, false, bosherr.WrapErrorf(err, "Generating variable '%s'", varDef.Name)
	}

	return val, true, nil
}

func (s VarsConfigServerStore) List() ([]boshtpl.VariableDefinition, error) {
	return nil, bosherr.Errorf("Listing of variables in config server is not supported")
}
