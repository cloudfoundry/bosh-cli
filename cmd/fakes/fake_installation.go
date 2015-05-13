package fakes

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	biinstallation "github.com/cloudfoundry/bosh-init/installation"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type FakeInstallation struct {
}

func (f *FakeInstallation) Target() biinstallation.Target {
	return biinstallation.Target{}
}

func (f *FakeInstallation) Job() biinstalljob.InstalledJob {
	return biinstalljob.InstalledJob{}
}

func (f *FakeInstallation) WithRunningRegistry(logger boshlog.Logger, stage biui.Stage, fn func() error) error {
	return fn()
}

func (f *FakeInstallation) StartRegistry() error {
	return nil
}

func (f *FakeInstallation) StopRegistry() error {
	return nil
}
