package installation

import (
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Installation

type Installation interface {
	Target() Target
	Jobs() []InstalledJob
}

type installation struct {
	target   Target
	jobs     []InstalledJob
	manifest biinstallmanifest.Manifest
}

func NewInstallation(
	target Target,
	jobs []InstalledJob,
	manifest biinstallmanifest.Manifest,
) Installation {
	return &installation{
		target:   target,
		jobs:     jobs,
		manifest: manifest,
	}
}

func (i *installation) Target() Target {
	return i.target
}

func (i *installation) Jobs() []InstalledJob {
	return i.jobs
}
