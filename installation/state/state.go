package state

import (
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
)

type State interface {
	RenderedCPIJob() biinstalljob.RenderedJobRef
	CompiledPackages() []biinstallpkg.CompiledPackageRef
}

type state struct {
	renderedCPIJob   biinstalljob.RenderedJobRef
	compiledPackages []biinstallpkg.CompiledPackageRef
}

func NewState(renderedCPIJob biinstalljob.RenderedJobRef, compiledPackages []biinstallpkg.CompiledPackageRef) State {
	return state{
		renderedCPIJob:   renderedCPIJob,
		compiledPackages: compiledPackages,
	}
}

func (s state) RenderedCPIJob() biinstalljob.RenderedJobRef {
	return s.renderedCPIJob
}

func (s state) CompiledPackages() []biinstallpkg.CompiledPackageRef {
	return s.compiledPackages
}
