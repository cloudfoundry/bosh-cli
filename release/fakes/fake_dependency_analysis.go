package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeDependencyAnalysis struct {
	DeterminePackageCompilationOrderResult  []*bmrel.Package
	DeterminePackageCompilationOrderRelease bmrel.Release
}

func NewFakeDependencyAnalysis() *FakeDependencyAnalysis {
	return &FakeDependencyAnalysis{}
}

func (fda *FakeDependencyAnalysis) DeterminePackageCompilationOrder(release bmrel.Release) ([]*bmrel.Package, error) {
	fda.DeterminePackageCompilationOrderRelease = release
	return fda.DeterminePackageCompilationOrderResult, nil
}
