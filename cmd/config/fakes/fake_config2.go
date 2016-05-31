package fakes

import (
	"github.com/cloudfoundry/bosh-init/cmd/config"
)

type FakeConfig2 struct {
	Existing ConfigContents

	Saved   *ConfigContents
	SaveErr error
}

type ConfigContents struct {
	TargetURL    string
	TargetAlias  string
	TargetCACert string

	Called bool
}

func (f *FakeConfig2) Target() string { return f.Existing.TargetURL }

func (f *FakeConfig2) Targets() []config.Target {
	panic("Not implemented")
}

func (f *FakeConfig2) ResolveTarget(targetOrName string) string {
	return ""
}

func (f *FakeConfig2) SetTarget(target, alias, caCert string) config.Config {
	f.Saved = &ConfigContents{}

	return &FakeConfig2{
		Existing: ConfigContents{
			TargetURL:    target,
			TargetAlias:  alias,
			TargetCACert: caCert,
		},

		Saved:   f.Saved,
		SaveErr: f.SaveErr,
	}
}

func (f *FakeConfig2) CACert(target string) string {
	return f.Existing.TargetCACert
}

func (f *FakeConfig2) Credentials(target string) config.Creds {
	panic("Not implemented")
}

func (f *FakeConfig2) SetCredentials(target string, creds config.Creds) config.Config {
	panic("Not implemented")
}

func (f *FakeConfig2) UnsetCredentials(target string) config.Config {
	panic("Not implemented")
}

func (f *FakeConfig2) Deployment(target string) string {
	panic("Not implemented")
}

func (f *FakeConfig2) SetDeployment(target string, nameOrPath string) config.Config {
	panic("Not implemented")
}

func (f *FakeConfig2) Save() error {
	f.Saved.TargetURL = f.Existing.TargetURL
	f.Saved.TargetAlias = f.Existing.TargetAlias
	f.Saved.TargetCACert = f.Existing.TargetCACert
	f.Saved.Called = true
	return f.SaveErr
}
