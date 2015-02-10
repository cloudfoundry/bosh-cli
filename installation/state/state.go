package state

import (
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
)

type State interface {
	RenderedCPIJob() bminstalljob.RenderedJobRef
	CompiledPackages() []bminstallpkg.CompiledPackageRef
}

type state struct {
	renderedCPIJob   bminstalljob.RenderedJobRef
	compiledPackages []bminstallpkg.CompiledPackageRef
}

func NewState(renderedCPIJob bminstalljob.RenderedJobRef, compiledPackages []bminstallpkg.CompiledPackageRef) State {
	return state{
		renderedCPIJob:   renderedCPIJob,
		compiledPackages: compiledPackages,
	}
}

func (s state) RenderedCPIJob() bminstalljob.RenderedJobRef {
	return s.renderedCPIJob
}

func (s state) CompiledPackages() []bminstallpkg.CompiledPackageRef {
	return s.compiledPackages
}
