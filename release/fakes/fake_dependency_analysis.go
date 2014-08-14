package fakes

import (
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeDependencyAnalysis struct {
	DeterminePackageCompilationOrderResult  []*bmrelease.Package
	DeterminePackageCompilationOrderRelease bmrelease.Release
}

func NewFakeDependencyAnalysis() *FakeDependencyAnalysis {
	return &FakeDependencyAnalysis{}
}

func (fda *FakeDependencyAnalysis) DeterminePackageCompilationOrder(release bmrelease.Release) ([]*bmrelease.Package, error) {
	fda.DeterminePackageCompilationOrderRelease = release
	return fda.DeterminePackageCompilationOrderResult, nil
}
