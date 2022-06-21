package cmdfakes

import (
	biinstallation "github.com/cloudfoundry/bosh-cli/v7/installation"
)

type FakeInstallation struct {
}

func (f *FakeInstallation) Target() biinstallation.Target {
	return biinstallation.Target{}
}

func (f *FakeInstallation) Job() biinstallation.InstalledJob {
	return biinstallation.InstalledJob{}
}
