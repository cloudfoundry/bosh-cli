package installation

import (
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
)

type Installation interface {
	Target() Target
	Job() InstalledJob
}

type installation struct {
	target   Target
	job      InstalledJob
	manifest biinstallmanifest.Manifest
}

func NewInstallation(
	target Target,
	job InstalledJob,
	manifest biinstallmanifest.Manifest,
) Installation {
	return &installation{
		target:   target,
		job:      job,
		manifest: manifest,
	}
}

func (i *installation) Target() Target {
	return i.target
}

func (i *installation) Job() InstalledJob {
	return i.job
}
